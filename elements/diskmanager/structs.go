package diskmanager

import "encoding/binary"

//Partition is a struct used to create partitions on a new mbr
type Partition struct {
	Name   [16]byte
	Type   byte
	Fit    byte
	Status byte
	Start  uint64
	Size   uint64
}

//Mbr is a struct used on mkdisk
type Mbr struct {
	Size          uint64
	CreationDate  [20]byte
	DiskSignature uint64
	Partitions    [4]Partition
}

//Ebr is a struct used to create extended and logic partitions
type Ebr struct {
	Status byte
	Fit    byte
	Start  uint64
	Size   uint64
	Next   int64
	Name   [16]byte
}

//EBRSIZE total size of the Ebr
var EBRSIZE = binary.Size(Ebr{})

//TotalEbrSize returns the totalSize of a ebr
func TotalEbrSize(sizeOnBytes uint64) uint64 {
	return uint64(EBRSIZE) + sizeOnBytes
}

//MountedDisk is a struct used to represent a mounted disk on mount
type MountedDisk struct {
	Path              string
	ID                byte //letra
	MountedPartitions []MountedPartition
	UsedIDs           []byte
}

//MountedPartition is a struct used to represent a mounted partition on mount
type MountedPartition struct {
	ID          string //el mero mero identificador
	Name        string //nombre real de la particion
	PartitionID byte   //numero auxiliar para identificarlo
}