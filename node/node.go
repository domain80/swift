package node

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"swift/global"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
)

type Node struct {
	basePort          int
	uiDir             embed.FS
	infoLog           *log.Logger
	errLog            *log.Logger
	connectionPool    map[int]net.Listener
	uiSocket          *websocket.Conn
	senderTimeout     time.Duration
	serverPort        int
	hostname          string
	backendConnection net.Conn
}

type Status struct {
	Status string `json:"status"`
}
type Intro struct {
	Status          string `json:"status"`
	Hostname        string `json:"hostname"`
	Conns           []int  `json:"connectionPool"`
	ConnectedNodeIP string `json:"connectedIP"`
}

func NewNode(infoLog *log.Logger, errLog *log.Logger, uiDir embed.FS) *Node {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("swift%d", rand.Intn(500))
	}

	return &Node{
		basePort:       4009,
		uiDir:          uiDir,
		infoLog:        infoLog,
		errLog:         errLog,
		connectionPool: make(map[int]net.Listener),
		senderTimeout:  time.Second * 20,
		hostname:       hostname,
	}
}

func (n *Node) Start() {
	// setup connection pool
	for i := 0; i < 5; i++ {
		port := n.getAvailablePort()
		n.connectionPool[port] = nil
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			n.errLog.Println(err)
			return
		}
		n.connectionPool[port] = listener
		go func() {
			fileRouter := chi.NewRouter()
			fileRouter.Use(middleware.Logger)
			fileRouter.HandleFunc("/{chunkSize}-{totalChunks}-{fileName}", n.handleFileReception)

			http.Serve(listener, fileRouter)
		}()
	}
	fmt.Printf("%+v\n", n.connectionPool)

	// begin UI server
	func() {
		n.infoLog.Println("Starting UI server")
		port := n.getAvailablePort()
		r := chi.NewRouter()
		r.Use(middleware.Logger)

		uiDir, err := fs.Sub(n.uiDir, "ui")
		if err != nil {
			n.errLog.Fatal("Unable to access uiDirectory")
		}

		uiStaticDir, err := fs.Sub(n.uiDir, "ui/static")
		if err != nil {
			n.errLog.Fatal("Unable to access uiDirectory")
		}

		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(uiStaticDir)))
		r.Handle("/*", http.FileServerFS(uiDir))
		r.HandleFunc("/sender", n.handleSenderRole)
		r.HandleFunc("/receiver", n.handleReceiverRole)
		n.infoLog.Printf("Server started on port %d\n", port)
		OpenPage(fmt.Sprintf("http://localhost:%d", port))
		http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	}()
}

func (n *Node) handleSenderRole(w http.ResponseWriter, r *http.Request) {
	// server role assumed
	// upgrade to websocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	n.uiSocket = conn
	conn.WriteJSON(struct {
		Status string `json:"status"`
	}{
		Status: "sender",
	})
	n.infoLog.Println(`{"status": "sender"}`)

	go n.ReadLoop(n.uiSocket)()

	// broadcast
	n.serverPort = n.getAvailablePort()
	go func() {
		n.broadcast()
	}()

	n.infoLog.Println("backend server started on port ", n.serverPort)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.serverPort))
	if err != nil {
		n.errLog.Println(listener)
		return
	}
	timer := time.AfterFunc(n.senderTimeout, func() {
		n.infoLog.Println("status", "Server timeout... shutting down")
		// close backend listener
		listener.Close()
	})

	n.backendConnection, err = listener.Accept()
	if err != nil {
		n.errLog.Println("accepting connection err: ", err)
	}
	timer.Stop()

	if n.backendConnection != nil {
		// Send an introduction message
		n.sendIntroduction()

		intro, err := n.receiveIntroduction()
		if err != nil {
			n.errLog.Println(err)
			return
		}
		n.infoLog.Printf("Received intro message from %s with numbers %v\n", intro.Hostname, intro.Conns)
		n.uiSocket.WriteJSON(intro)
		go n.ReadLoop(n.backendConnection)()
	} else {
		n.uiSocket.WriteJSON(Status{Status: "server timed out; no connections made"})
	}
}

func (n *Node) handleReceiverRole(w http.ResponseWriter, r *http.Request) {
	// receiver role assumed
	// upgrade to websocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		n.errLog.Println(err)
		return
	}
	n.uiSocket = conn
	conn.WriteJSON(Status{
		Status: "receiver",
	})
	n.infoLog.Println(`{"status": "receiver"}`)

	go n.ReadLoop(n.uiSocket)()

	// start listening
	availableHost, err := n.Listen()
	if err != nil {
		n.errLog.Println(err)
		return
	}
	err = n.Connect(availableHost)
	if err != nil {
		n.errLog.Println(err)
		return
	}

	intro, err := n.receiveIntroduction()
	if err != nil {
		n.errLog.Println(err)
		return
	}
	n.sendIntroduction()
	n.infoLog.Printf("Received intro message from %s with numbers %v\n", intro.Hostname, intro.Conns)
	n.uiSocket.WriteJSON(intro)

	go n.ReadLoop(n.backendConnection)()
}

func (n *Node) Listen() (string, error) {
	// Resolve the broadcast address and port
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", BroadcastPort))
	if err != nil {
		return "", err
	}

	// Create a UDP socket to listen on
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Set a timeout for the socket
	conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	n.infoLog.Println("now listening on port ", global.BroadcastPort)

	// Wait for a message
	stopTime := time.Now().Add(n.senderTimeout)
	availableHost := ""
	for time.Now().Before(stopTime) {
		buffer := make([]byte, 40)
		l, remoteAddr, err := conn.ReadFromUDP(buffer)
		// c.AvailableHosts = append(c.AvailableHosts, Host{hostname: "", ipPort: })
		if err != nil {
			return "", err
		}

		n.infoLog.Println("broadcast received: ", remoteAddr.String(), string(buffer))
		availableHost = fmt.Sprintf("%s:%s",
			strings.Split(remoteAddr.String(), ":")[0],
			strings.Split(string(buffer[:l]), ":")[1],
		)
		if availableHost != "" {
			break
		}
	}
	return availableHost, nil
}

func (n *Node) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	n.backendConnection = conn

	// send hostname
	return nil
}

func (n *Node) ReadLoop(conn interface{}) func() {
	switch v := conn.(type) {
	case net.Conn:
		return func() {
			for {
				readBytes := new(bytes.Buffer)
				_, err := v.Read(readBytes.Bytes())
				if err != nil {
					n.errLog.Println(err)
					break
				}

				if readBytes.String() != "" {
					n.infoLog.Println(readBytes.String())
				}
			}
		}
	case *websocket.Conn:
		return func() {
			for {
				_, content, err := v.ReadMessage()
				if err != nil {
					n.errLog.Println(err)
					break
				}

				n.infoLog.Println(content)
			}
		}
	default:
		return nil
	}
}
