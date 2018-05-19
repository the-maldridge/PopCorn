package ctl

var (
	server string
	port   int
)

func SetServer(s string) { server = s }
func SetPort(p int)      { port = p }
