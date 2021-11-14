package main

import (
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
	Type string
	Data string
}

func (e *IPCMessage) String() string {
	x, _ := json.MarshalIndent(e, "  ", "  ")
	return string(x)
}

func (e *IPCMessage) Write(c net.Conn) error {
	if err := json.NewEncoder(c).Encode(e); err != nil {
		return err
	}
	if _, err := c.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func (e *IPCMessage) Read(c net.Conn) error {
	decoder := json.NewDecoder(c)
	for decoder.More() {
		if err := decoder.Decode(e); err != nil {
			return err
		}
	}
	return nil
}
