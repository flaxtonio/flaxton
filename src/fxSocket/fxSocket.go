package fxSocket

import (
	"net"
	"encoding/json"
	"encoding/binary"
	"io"
)


type Socket struct {
	Conn *net.TCPConn
}

type SocketMessage struct {
	Event string `json:"event"` // event name to be fired
	Data []byte  `json:"data"` // Data in JSON format
}

func (sm *SocketMessage) ToBytes() (ret_data []byte, err error) {
	ret_data, err = json.Marshal(sm)
	if err != nil {
		return
	}
	len_byte := make([]byte, 4)
	binary.LittleEndian.PutUint32(len_byte, uint32(len(ret_data)))
	ret_data = append(len_byte, ret_data...)
	return
}

func (sm *SocketMessage) FromBytes(data []byte) (err error) {
	err = json.Unmarshal(data, sm)
	return
}


func (s *Socket) ReadData() (message SocketMessage, err error) {
	var (
		dlen uint32 // read data length
		len_byte = make([]byte, 4) // putting data length here
	)

	_, err = s.Conn.Read(len_byte)
	if err != nil {
		return
	}
	dlen = binary.LittleEndian.Uint32(len_byte)
	core_data := make([]byte, dlen)
	_, err = s.Conn.Read(core_data)
	// TODO: Handle if there is content smaller than dlen
	if err != nil {
		return
	}
	// Filling data to message
	message.FromBytes(core_data)
	return
}

func (s *Socket) WriteData(message SocketMessage) (err error) {
	var byte_data []byte
	byte_data, err = message.ToBytes()
	if err != nil {
		return
	}
	_, err = s.Conn.Write(byte_data)
	return
}

func (s *Socket) Emit(event string, data interface{}) (err error) {
	var (
		message SocketMessage
	)
	message.Data, err = json.Marshal(data)
	if err != nil {
		return
	}
	message.Event = event
	s.WriteData(message)
	return
}

type SocketComponent struct {
	OnError func(error)
	OnConnection func(Socket)
	OnDisconnect func(Socket)
	Events map[string][]func([]byte, Socket)
}

func (sc *SocketComponent) On(event string, f func([]byte, Socket)) {
	sc.Events[event] = append(sc.Events[event], f)
}

func (sc *SocketComponent) HandleEventMessage(message SocketMessage, s Socket) {
	if _, ok := sc.Events[message.Event]; !ok {
		return
	}

	for _, f :=range sc.Events[message.Event] {
		f(message.Data, s)
	}
}

type Child struct {
	SocketComponent
	ParentConnection Socket
	ParentAddress string
}

type Parent struct {
	SocketComponent
	Listener *net.TCPListener
	Childs map[string]Socket
}

func NewParent(address string) (p Parent, err error) {
	var tcp_addr *net.TCPAddr
	tcp_addr, err = net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return
	}
	p.Listener, err = net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		return
	}
	p.Events = make(map[string][]func([]byte, Socket))
	p.Childs = make(map[string]Socket)
	// Just for not checking function for nil every time
	p.OnError = func(e error) {}
	p.OnConnection = func(s Socket) {}
	p.OnDisconnect = func(s Socket) {}
	return
}

func (sc *SocketComponent) HandleConnection(s Socket) {
	var (
		message SocketMessage
		err error
	)

	sc.OnConnection(s)

	for {
		message, err = s.ReadData()
		if err != nil {
			if err != io.EOF {
				sc.OnError(err)
			}
			break
		}
		sc.HandleEventMessage(message, s)
	}

	sc.OnDisconnect(s)
}

func (p *Parent) Listen() {
	var (
		conn *net.TCPConn
		err error
	)

	for {
		conn, err = p.Listener.AcceptTCP()
		if err != nil {
			p.OnError(err)
		}
		go p.HandleConnection(Socket{Conn:conn})
	}
}

func NewChild(parent_address string) (child Child, err error) {
	var (
		conn *net.TCPConn
		tcpaddr *net.TCPAddr
	)
	tcpaddr, err = net.ResolveTCPAddr("tcp", parent_address)
	if err != nil {
		return
	}
	conn, err = net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return
	}
	child.Events = make(map[string][]func([]byte, Socket))
	child.ParentConnection = Socket{Conn:conn}
	child.ParentAddress = parent_address
	child.OnError = func(e error) {}
	child.OnConnection = func(s Socket) {}
	child.OnDisconnect = func(s Socket) {}
	return
}

func (child *Child) Emit(event string, data interface{}) {
	child.ParentConnection.Emit(event, data)
}