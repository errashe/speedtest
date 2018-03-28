package Structs

import "time"

type Server struct {
	ID        int    `storm:"id,increment"`
	IP        string `storm:"unique"`
	Download  float64
	Upload    float64
	Ping      float64
	Timestamp time.Time `storm:"index"`
}
