package main

import (
	"fmt"
	"os"
	"path"
)

const SYNC_BYTE byte = 0x47 // MPEG-2 TS sync byte value

func seekTillFirstSyncByte(pFile *os.File) int64 {
	singleByte := make([]byte, 1)
	var off int64 = -1
	for n, err := pFile.Read(singleByte); n == 1 && err == nil; n, err = pFile.Read(singleByte) {
		if singleByte[0] == SYNC_BYTE {
			off, _ = pFile.Seek(-1, os.SEEK_CUR)
			break
		}
	}
	return off
}

func getTpSize(pFile *os.File, firstSyncByteOffset int64) int {
	singleByte := make([]byte, 1)
	pFile.Seek(firstSyncByteOffset, os.SEEK_SET)
	// Try 188
	pFile.Seek(188, os.SEEK_CUR)
	n, _ := pFile.Read(singleByte)
	if n == 1 && singleByte[0] == SYNC_BYTE {
		pFile.Seek(firstSyncByteOffset, os.SEEK_SET)
		return 188
	}
	// Not 188, maybe 204, try 15 bytes further
	pFile.Seek(15, os.SEEK_CUR)
	n, _ = pFile.Read(singleByte)
	if n == 1 && singleByte[0] == SYNC_BYTE {
		pFile.Seek(firstSyncByteOffset, os.SEEK_SET)
		return 204
	}

	// Reset offset to beginning of TP
	pFile.Seek(firstSyncByteOffset, os.SEEK_SET)
	return -1
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Forgot TS file name?\n")
		return
	}

	inFileName := os.Args[1]
	outFileName := path.Base(inFileName) + "-188" + path.Ext(inFileName)

	inFile, _ := os.OpenFile(inFileName, os.O_RDONLY, 0660)
	beginOfTPs := seekTillFirstSyncByte(inFile)
	if beginOfTPs < 0 {
		fmt.Printf("Unable to find any sync byte in the file\n")
		return
	}
	tpSize := getTpSize(inFile, beginOfTPs)
	if tpSize != 188 && tpSize != 204 {
		fmt.Printf("Could not determine if proper TP size, maybe the file is not a TS?\n")
		return
	} else if tpSize == 188 {
		fmt.Printf("Input file is already written in 188-byte format, not doing anything\n")
	} else {
		fmt.Printf("Creating file %s\n", outFileName)
		outFile, err := os.OpenFile(outFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)
		if err != nil {
			fmt.Printf("Could not create output file: %s\n", err)
			return
		}
		tp204 := make([]byte, 204)
		nConvertedPackets := 0
		for n, err := inFile.Read(tp204); n == 204 && err == nil; n, err = inFile.Read(tp204) {
			outFile.Write(tp204[0:188])
			nConvertedPackets += 1
		}
		fmt.Printf("Converted %d packets\n", nConvertedPackets)
	}
}
