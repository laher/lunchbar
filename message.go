package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
)

const (
	// start at 1 so that zero-messages are ignored
	msgPluginID               = "plid"
	msgPluginRefreshAll       = "real"
	msgPluginRefreshComplete  = "recm"
	msgPluginRefreshError     = "reer"
	msgPluginUnrecognised     = "unre"
	msgPluginQuit             = "quit"
	msgPluginRestartme        = "reme"
	msgSupervisorRefresh      = "refr"
	msgSupervisorUnrecognised = "unre"
	msgSupervisorQuit         = "quit"
)

type IPCMessage struct {
	Length int
	Type   string
	Data   string
}

func (e *IPCMessage) String() string {
	x, _ := json.MarshalIndent(e, "  ", "  ")
	return string(x)
}

func (e *IPCMessage) Write(c net.Conn) error {
	e.Length = len(e.Data)
	data := make([]byte, 0, 8+e.Length)

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(e.Length))
	data = append(data, buf...)

	data = append(data, []byte(e.Type)...)

	/* w := bytes.Buffer{}
	if err := binary.Write(&w, binary.BigEndian, e.Data); err != nil {
		return err
	}
	*/
	data = append(data, []byte(e.Data)...)
	if _, err := c.Write(data); err != nil {
		return err
	}

	return nil
}

func (e *IPCMessage) Read(c net.Conn) error {
	buf := make([]byte, 4)
	if _, err := c.Read(buf); err != nil {
		return err
	}
	byteCount := binary.BigEndian.Uint32(buf)
	e.Length = int(byteCount)

	buf = make([]byte, 4)
	if _, err := c.Read(buf); err != nil {
		return err
	}
	e.Type = string(buf)

	data := make([]byte, e.Length)

	if _, err := c.Read(data); err != nil {
		return err
	}
	e.Data = string(data)

	return nil
}
