package main

import (
	"log"

	"github.com/dualm/melsec"
)

func main() {
	conn, err := melsec.NewConn("192.168.0.2", "7000", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	svDevice := melsec.NewMultiDevice(conn)

	svDevice.AddBlock("D12000", 20) // 0
	svDevice.AddBlock("D12100", 22) // 1
	svDevice.AddBlock("D12200", 6)  // 2
	svDevice.AddBlock("D12300", 16) // 3
	svDevice.AddBlock("D1270", 20)  // 4
	svDevice.AddBlock("R102", 36)   // 5
	svDevice.AddBlock("R190", 22)   // 6
	svDevice.AddBlock("R220", 18)   // 7
	svDevice.AddBlock("R320", 18)   // 8
	svDevice.AddBlock("R270", 8)    // 9
	svDevice.AddBlock("R450", 6)    // 10
	svDevice.AddBlock("R485", 5)    // 11
	svDevice.AddBlock("M20000", 6)  // 12
	svDevice.AddBlock("M20100", 1)  // 13
	svDevice.AddBlock("W4000", 69)  // 14
	svDevice.AddBlock("X800", 18)   // 15
	svDevice.AddBlock("D1600", 20)  // 16

	if err := svDevice.Read(); err != nil {
		log.Fatal(err)
	}

	v := svDevice.GetValue()

	for i := range v {
		log.Printf("% x", v[i])
	}

	
}
