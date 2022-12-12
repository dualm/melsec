package main

import (
	"log"
	"time"

	"github.com/dualm/melsec"
)

func main() {
	conn, err := melsec.NewConn("192.168.0.11", "7000", nil, melsec.SetCPUTimer(uint16(10)))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	log.Println(conn.GetCPUInfo())

	svDevice0, err := melsec.NewDevice("D12000", 480, conn)
	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()
	if err := svDevice0.Read(); err == nil {
		svDevice0.GetValue()
	}
	log.Println(time.Since(t).Milliseconds(), "ms")

	svDevice1, err := melsec.NewMultiDevice(conn)
	if err != nil {
		log.Fatal(err)
	}

	svDevice2, err := melsec.NewMultiDevice(conn)
	if err != nil {
		log.Fatal(err)
	}

	// 添加数据区块
	svDevice1.AddBlock("D12000", 48) // 0
	svDevice1.AddBlock("D12100", 12) // 1
	svDevice1.AddBlock("D12200", 6)  // 2
	svDevice1.AddBlock("D12300", 34) // 3
	svDevice1.AddBlock("D12400", 24) // 4
	svDevice1.AddBlock("D1270", 20)  // 5
	svDevice1.AddBlock("R102", 10)   // 6

	svDevice2.AddBlock("R320", 2)    // 7
	svDevice2.AddBlock("R450", 8)    // 8
	svDevice2.AddBlock("M20000", 10) // 9
	svDevice2.AddBlock("W1000", 8)   // 10
	svDevice2.AddBlock("X800", 4)    // 11
	svDevice2.AddBlock("D1600", 2)   // 12
	svDevice2.AddBlock("W651B", 2)   // 13
	svDevice2.AddBlock("D600", 2)    // 17
	svDevice2.AddBlock("D670", 2)    // 18

	devices := []*melsec.MultiDevice{svDevice1, svDevice2}

	re := make([][]byte, 0)

	t = time.Now()
	for i := range devices {
		if err := devices[i].Read(); err != nil {
			log.Fatal(err)
		}

		re = append(re, devices[i].GetValue()...)
		log.Println("Multi: ", i)
	}
	log.Println(time.Since(t).Milliseconds(), "ms")

	svDevice, err := melsec.NewDevice("D12000", 1, conn)
	if err != nil {
		log.Fatal(err)
	}

	t = time.Now()
	err = svDevice.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(time.Since(t).Milliseconds(), "ms")
}
