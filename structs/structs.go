package Structs

import (
	. "fmt"
	"time"
)

type Server struct {
	ID        int    `storm:"id,increment"`
	IP        string `storm:"unique"`
	Download  float64
	Upload    float64
	Ping      float64
	Timestamp time.Time `storm:"index"`
}

func (s Server) String() string {
	return Sprintf("%-15s | %8.2f/%8.2f mbit/s | %7.2f ms | %10s",
		s.IP,
		s.Download,
		s.Upload,
		s.Ping,
		s.Timestamp.Format("02-01-2006 15:04:05"),
	)
}

type History struct {
	ID        int    `storm:"id,increment"`
	IP        string `storm:"index"`
	Download  float64
	Upload    float64
	Ping      float64
	Timestamp time.Time `storm:"index"`
}

func (h *History) Copy(s Server) {
	h.IP = s.IP
	h.Download = s.Download
	h.Upload = s.Upload
	h.Ping = s.Ping
	h.Timestamp = s.Timestamp
}

type Command struct {
	Command string `json:"command"`
	Value   string `json:"value"`
}

type Quer struct {
	IP    string `query:"ip"`
	Count int    `query:"count"`
}
