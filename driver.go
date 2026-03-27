package web

import (
	"net/http"

	. "github.com/infrago/base"
)

type (
	// Driver defines web driver interface.
	Driver interface {
		Connect(*Instance) (Connection, error)
	}

	// Connection defines web connection interface.
	Connection interface {
		Open() error
		Close() error

		Register(name string, info Info, hosts []string) error
		Upgrade(res http.ResponseWriter, req *http.Request) (Socket, error)

		Start() error
		StartTLS(certFile, keyFile string) error
	}

	Socket interface {
		ReadMessage() (int, []byte, error)
		WriteMessage(messageType int, data []byte) error
		Close() error
		Raw() Any
	}

	// Delegate handles web requests.
	Delegate interface {
		Serve(name string, params Map, res http.ResponseWriter, req *http.Request)
	}

	// Info contains route information.
	Info struct {
		Method string
		Uri    string
		Router string
		Entry  string
		Args   Vars
	}
)
