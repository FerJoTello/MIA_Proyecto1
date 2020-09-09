package diskmanager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

//MountedDisks is an array that contains all the information of mount
var MountedDisks []MountedDisk

//Mkdisk function
func Mkdisk(path string, name string, unit string, size int) {
	if size > 0 && path != "" && name != "" {
		var sizeOnBytes uint64
		if unit != "k" {
			//units on megabytes (1024*1024)
			sizeOnBytes = 1024 * 1024 * uint64(size)
		} else {
			//units on kylobytes (1024)
			sizeOnBytes = 1024 * uint64(size)
		}
		// partitions initialization
		var partitions [4]Partition
		for i := 0; i < 4; i++ {
			copy(partitions[i].Name[:], "new_part")
			partitions[i].Status = 'F'
			partitions[i].Fit = 'x'
			partitions[i].Type = 'x'
			partitions[i].Start = 0
			partitions[i].Size = 0
		}
		// disk init
		newMbr := Mbr{}
		newMbr.Size = sizeOnBytes
		newMbr.Partitions = partitions
		newMbr.DiskSignature = uint64(time.Now().Unix())
		copy(newMbr.CreationDate[:], time.Now().String()[0:19])
		writeMbr(path, name, sizeOnBytes, newMbr)
	} else {
		fmt.Println("No es posible ejecutar el comando \"mkdisk\". Parametros no definidos")
	}
}

//Rmdisk function
func Rmdisk(path string) {
	if path != "" {
		fmt.Println("Se eliminará el disco. Está seguro? (Y/N)")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			err := os.Remove(path)
			if err == nil {
				println("Se eliminó correctamente")
				return
			}
		}
	}
	fmt.Println("No se ha podido eliminar el disco. No existe")
}

//Fdisk function
func Fdisk(path string, name string, unit string, typee string, fit string, delete string, add int, size int) {
	//checking if the params are defined
	if name != "" && path != "" && size > 0 {
		if add == 0 && delete == "" { //does not add space and neither deletes anything so it should create a partition
			//reading file
			file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
			defer file.Close()
			if err != nil {
				fmt.Println("No se pudo abrir el archivo")
				return
			}
			mbr, err := canCreatePartition(file)
			if err != nil {
				if err.Error() != "particiones ya definidas. no puede crear más" {
					fmt.Println("No se puede crear particion", err.Error())
				}
				return
			}
			var nameOnBytes [16]byte
			copy(nameOnBytes[:], name)
			if existName(nameOnBytes, mbr, file) {
				fmt.Println("No se puede crear particion con un nombre ya existente")
				return
			}
			//Setting size
			var sizeOnBytes uint64
			if unit == "" || unit == "k" {
				//kilobytes
				sizeOnBytes = uint64(size) * 1024
			} else if unit == "b" {
				//bytes
				sizeOnBytes = uint64(size)
			} else if unit == "m" {
				//megabytes
				sizeOnBytes = uint64(size) * 1024 * 1024
			} else {
				fmt.Println("Está enviando un parámetro no válido para unit")
				return
			}
			//Setting fit
			fit = strings.ToUpper(fit)
			var fitOnByte byte
			if fit == "" { // Worst Fit by default
				fitOnByte = 'W'
			} else if fit == "BF" || fit == "FF" || fit == "WF" {
				fitOnByte = byte(fit[0])
			} else {
				fmt.Println("Tipo de ajuste indicado no valido")
				return
			}
			//setting type and executing creation
			var typeOnByte byte
			if strings.ToLower(typee) == "e" { //extended partition
				typeOnByte = 'E'
				createExtendedPartition(file, mbr, nameOnBytes, sizeOnBytes, fitOnByte, typeOnByte)
			} else if strings.ToLower(typee) == "l" { //logic partition
				typeOnByte = 'L'
				createLogicPartition(file, mbr, nameOnBytes, fitOnByte, sizeOnBytes)
			} else if strings.ToLower(typee) == "p" || typee == "" { //primary partition
				typeOnByte = 'P'
				createPrimaryPartition(file, mbr, nameOnBytes, sizeOnBytes, fitOnByte, typeOnByte)
			} else {
				fmt.Println("No se ha podido crear la particion. El tipo indicado no es valido.")
				return
			}
		}
	}
}

//Mount function
func Mount(path string, name string) {
	//checking if the params are defined
	if path == "" && name == "" {
		if len(MountedDisks) == 0 {
			fmt.Println("No hay discos montados")
		} else {
			for i := 0; i < len(MountedDisks); i++ {
				for j := 0; j < len(MountedDisks[i].MountedPartitions); j++ {
					fmt.Println("-id->", MountedDisks[i].MountedPartitions[j].ID, "-path->", MountedDisks[i].Path, "-name->", MountedDisks[i].MountedPartitions[j].Name)
				}
			}
		}
	} else if path != "" && name != "" {
		//validations
		file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm) //opens path's file
		defer file.Close()
		if err != nil {
			fmt.Println("No se pudo abrir el archivo")
			return
		}
		mbr, err := readDsk(file) //obtains mbr
		if err != nil {
			fmt.Println("No se pudo recuperar la info del mbr.")
			return
		}
		var i int
		for i = 0; i < len(MountedDisks); i++ { //finds disk index
			if MountedDisks[i].Path == path { //disk has an id
				break
			}
		}
		var disk MountedDisk
		if i == len(MountedDisks) { //disk doesn't have an id
			disk.ID = 97 + byte(i)
			disk.Path = path
			MountedDisks = append(MountedDisks, disk)
		} else {
			disk = MountedDisks[i]
		}
		var nameOnBytes [16]byte
		copy(nameOnBytes[:], name)
		if !existName(nameOnBytes, mbr, file) { //checks if the partition with the provided name exists
			fmt.Println("No existe el nombre de esa particion")
			return
		}
		for j := 0; j < len(disk.MountedPartitions); j++ { //checks if the partition already was defined
			if disk.MountedPartitions[j].Name == name {
				fmt.Println("Particion ya montada")
				return
			}
		}
		var partID int
		for partID = 0; partID < len(disk.UsedIDs); partID++ {
			if disk.UsedIDs[partID] == 0 { //id available
				break
			}
		}
		newMP := MountedPartition{} //"mounts" a new partition
		newMP.PartitionID = 49 + byte(partID)
		newMP.ID = "vd" + string(disk.ID) + string(newMP.PartitionID)
		newMP.Name = name
		disk.MountedPartitions = append(disk.MountedPartitions, newMP)
		disk.UsedIDs = append(disk.UsedIDs, newMP.PartitionID)
		MountedDisks[i] = disk
	} else {
		fmt.Println("Faltan parametros para el comando mount")
		return
	}
	fmt.Println("Fin comando mount")
}

//Unmount function
func Unmount(ID string) {
	if ID != "" {
		var disk MountedDisk
		//finding disk
		var diskIndex int
		diskID := ID[2]
		for diskIndex = 0; diskIndex < len(MountedDisks); diskIndex++ {
			if MountedDisks[diskIndex].ID == diskID {
				disk = MountedDisks[diskIndex]
				break
			}
		}
		//disk founded. now find partition
		for i, partID := 0, ID[3]; i < len(disk.MountedPartitions); i++ {
			if disk.MountedPartitions[i].PartitionID == partID {
				//partition founded. now has to be "unmounted"
				copy(disk.MountedPartitions[i:], disk.MountedPartitions[i+1:])
				disk.MountedPartitions = disk.MountedPartitions[:len(disk.MountedPartitions)-1]
				disk.UsedIDs[i] = 0
				break
			}
		}
		//partition "unmounted". now check size
		//now update
		MountedDisks[diskIndex] = disk
		/*
			if len(disk.MountedPartitions)==0{

			}*/
	} else {
		fmt.Println("Falta parametro para \"unmount\"")
	}
	fmt.Println("Fin comando unmount")
}

func writeMbr(path string, fileName string, sizeOnBytes uint64, newMbr Mbr) {
	createFolders(path)
	//the file is created with the provided name and path
	file, err := os.Create(path + "/" + fileName)
	//escribir en archivo
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	var cero int8 = 0
	s := &cero
	//Primer binario para iniciar escribiendo el 0 como inicio del archivo.
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, s)
	escribirBytes(file, binario.Bytes())
	//Posicionar en la ultima posicion-1 para cumplir con el tamano. Se escribe un 0
	file.Seek(int64(sizeOnBytes-1), 0) // segundo parametro: 0, 1, 2.     0 -> Inicio, 1-> desde donde esta el puntero, 2 -> Del fin para atras
	//Segundo Binario para definir el tamanio del archivo
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, s)
	escribirBytes(file, binario2.Bytes())
	//Se escribe el struct en el inicio del archivo
	file.Seek(0, 0)
	s1 := &newMbr
	//binario para escribir en el archivo creado con el tamanio y con el struct definido
	var binario3 bytes.Buffer
	binary.Write(&binario3, binary.BigEndian, s1)
	escribirBytes(file, binario3.Bytes())
	//fmt.Println("File created succesfully xd")
}

func escribirBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func createFolders(path string) {
	//checking if the path contains quotation marks
	if strings.Index(path, "\"") == 0 && strings.LastIndex(path, "\"") == len(path)-1 {
		path = strings.ReplaceAll(path, "\"", "")
	}
	//the path should exist
	if !existFolders(path) {
		//if doesn't exist it's created
		mkdirErr := os.MkdirAll(path, os.ModePerm)
		if mkdirErr != nil {
			log.Fatal(mkdirErr)
		}
		//fmt.Println("Check path. Should exist by now")
	}
}

func existFolders(path string) bool {
	if _, err := os.Stat(path); os.IsExist(err) {
		//exists
		return true
	}
	return false
}

func createExtendedPartition(file *os.File, mbr Mbr, partName [16]byte, size uint64, fit byte, typee byte) {
	//creates a new partition with the provided name, size and the worst adjustment
	newPartition := Partition{}
	newPartition.Name = partName
	newPartition.Status = 'F'
	newPartition.Fit = fit
	newPartition.Type = typee
	newPartition.Size = size
	//newPartition.Start is not defined yet. should find where starts the next available space
	var i int //"i" is the partition index
	newPartition.Start, i = getStartPartition(size, mbr)
	if i != -1 { //there was enough space for the new partition
		mbr.Partitions[i] = newPartition
	} else {
		fmt.Println("!No hay suficiente espacio en el disco para crear una nueva particion!")
		return
	}
	//the partitions are ordered
	mbr = orderPartitionsByStart(mbr)
	//the disk is updated
	file.Seek(0, 0)
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, &mbr)
	escribirBytes(file, binario.Bytes())
	//and the ebr info is placed on the disk
	ebr := Ebr{Start: newPartition.Start, Next: -1}
	file.Seek(int64(ebr.Start), 0)
	binario.Reset()
	binary.Write(&binario, binary.BigEndian, &ebr)
	escribirBytes(file, binario.Bytes())
}

func createPrimaryPartition(file *os.File, mbr Mbr, partName [16]byte, size uint64, fit byte, typee byte) {
	//creates a new partition with the provided name, size and adjustment
	newPartition := Partition{}
	newPartition.Name = partName
	newPartition.Status = 'F'
	newPartition.Fit = fit
	newPartition.Type = typee
	newPartition.Size = size
	//newPartition.Start is not defined yet. should find where starts the next available space
	var i int //"i" is the partition index
	newPartition.Start, i = getStartPartition(size, mbr)
	if i != -1 { //there was enough space for the new partition
		mbr.Partitions[i] = newPartition
	} else {
		fmt.Println("!No hay suficiente espacio en el disco para crear una nueva particion!")
		return
	}
	//the partitions are ordered
	mbr = orderPartitionsByStart(mbr)
	//the disk is updated
	file.Seek(0, 0)
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, &mbr)
	escribirBytes(file, binario.Bytes())
}

func createLogicPartition(file *os.File, mbr Mbr, partName [16]byte, fit byte, size uint64) {
	//find if a extended partition exists
	var extendedIndex int
	var existsExtended bool
	for extendedIndex = 0; extendedIndex < 4; extendedIndex++ {
		if mbr.Partitions[extendedIndex].Type == 'E' {
			existsExtended = true
			break
		}
	}
	if !existsExtended {
		fmt.Println("No existe particion extendida donde pueda agregar una particion logica.")
		return
	}
	extended := mbr.Partitions[extendedIndex]
	//obtaining ebr
	actualPosition := int64(extended.Start)
	actualEbr := Ebr{}
	var ebrList []Ebr //contains founded ebr
	//there could exist more than one ebr partition on the extended so it's necessary to find where its logic ends
	for true {
		file.Seek(int64(actualPosition), 0)
		//tries to obtain a possibly ebr
		data := readNextBytes(file, int(EBRSIZE))
		buffer := bytes.NewBuffer(data)
		err := binary.Read(buffer, binary.BigEndian, &actualEbr)
		if err != nil { //couldn't obtain ebr
			fmt.Println("binary.Read failed", err)
			return
		}
		//ebr obtained
		ebrList = append(ebrList, actualEbr) //adding founded ebr on ebrList
		actualPosition = actualEbr.Next      //where is the next ebr (in case that exists)
		//-1 if the ebr doesn't have a next
		if actualPosition == -1 { //means that the last ebr was founded
			break
		}
	}
	//ebr is updated with the data provided
	actualEbr.Status = 'F'
	actualEbr.Fit = fit
	actualEbr.Size = size
	actualEbr.Next = -1
	actualEbr.Name = partName
	//actualEbr.Start is not defined yet since it's necessary to find its position
	if len(ebrList) > 1 { //more than one ebr
		for i := 0; i < len(ebrList); i++ { //looking through the ebrs to find the first fit
			if ebrList[i].Next == -1 { //the next position is available (ebrList[i] is the last ebr)
				if TotalEbrSize(actualEbr.Size) < (extended.Start+extended.Size)-(ebrList[i].Start+TotalEbrSize(ebrList[i].Size)) { //ebr size should be less than the free space (total partition space - occupied space by the last ebr)
					actualEbr.Start = ebrList[i].Start + TotalEbrSize(ebrList[i].Size)
					ebrList[i].Next = int64(actualEbr.Start)
					ebrList = append(ebrList, actualEbr)
					break
				}
			} else if ebrList[i].Next >= 0 { //next position is occupied
				if TotalEbrSize(actualEbr.Size) < (ebrList[i+1].Start+TotalEbrSize(ebrList[i+1].Size))-(ebrList[i].Start+TotalEbrSize(ebrList[i].Size)) { //checks if there is enough space to set the actualEbr in between another two ebrs
					actualEbr.Start = TotalEbrSize(ebrList[i].Size) + ebrList[i].Start
					ebrList[i].Next = int64(actualEbr.Start)
					actualEbr.Next = int64(ebrList[i+1].Start)
					ebrList = append(ebrList, actualEbr)
					break
				}
			}
			if i == len(ebrList) {
				fmt.Println("No hay espacio para el nuevo ebr")
				return
			}
		}
	} else if len(ebrList) == 1 { //first ebr
		if ebrList[0].Status == 0 { //first ebr is empty
			if TotalEbrSize(actualEbr.Size) < extended.Size { //ebr size should be less than extended partition total size
				actualEbr.Start = extended.Start
				ebrList[0] = actualEbr
			} else {
				fmt.Println("El tamano del nuevo ebr es mayor que la partición que lo contiene")
				return
			}
		} else { //First ebr is not empty
			if TotalEbrSize(actualEbr.Size) < extended.Size-TotalEbrSize(ebrList[0].Size) { //ebr size should be less than extended partition free space
				ebrList[0].Next = int64(ebrList[0].Start + TotalEbrSize(ebrList[0].Size))
				actualEbr.Start = ebrList[0].Start + TotalEbrSize(ebrList[0].Size)
				ebrList = append(ebrList, actualEbr)
			} else {
				fmt.Println("El tamano del nuevo ebr es mayor que el espacio disponible")
				return
			}
		}
	} else {
		fmt.Println("XD") //ni yo sé cómo entraría aquí pero por si acaso xd
		return
	}
	//logic partition created. ebrs are updated on the disk
	for i := 0; i < len(ebrList); i++ {
		println("Start:", int64(ebrList[i].Start))
		file.Seek(int64(ebrList[i].Start), 0)
		var binario bytes.Buffer
		binary.Write(&binario, binary.BigEndian, &ebrList[i])
		escribirBytes(file, binario.Bytes())
		binario.Reset()
	}
}

//obtains the new partition start position and the mbr partitions' index
func getStartPartition(size uint64, mbr Mbr) (uint64, int) {
	if mbr.Partitions[0].Size != 0 { //first is occupied
		for i := 1; i < 4; i++ {
			if mbr.Partitions[i].Size == 0 { //partition available at the "i" index
				//size of the new partition should be less than the total free space (free space = total space - where ends the last defined partition)
				if size < mbr.Size-(mbr.Partitions[i-1].Start-mbr.Partitions[i-1].Size) {
					return mbr.Partitions[i-1].Start + mbr.Partitions[i-1].Size, i
				}
			}
		}
	} else { //first is available
		if size < mbr.Size-uint64(binary.Size(mbr)) { //size of the new partition should be less than the free space
			return uint64(binary.Size(mbr)), 0
		}
	}
	return 0, -1
}

func orderPartitionsByStart(m Mbr) Mbr {
	for i := 0; i < len(m.Partitions); i++ {
		for j := 1; i < i; j++ {
			if m.Partitions[j].Start > m.Partitions[j+1].Start {
				aux := m.Partitions[j]
				m.Partitions[j] = m.Partitions[j+1]
				m.Partitions[j+1] = aux
			}
		}
	}
	for i := 0; i < len(m.Partitions); i++ {
		for j := 1; i < i; j++ {
			if m.Partitions[j].Start < m.Partitions[j+1].Start && m.Partitions[j].Start == 0 {
				aux := m.Partitions[j]
				m.Partitions[j] = m.Partitions[j+1]
				m.Partitions[j+1] = aux
			}
		}
	}
	return m
}

func canCreatePartition(file *os.File) (Mbr, error) {
	//obtaining mbr
	mbr, err := readDsk(file)
	if err != nil {
		fmt.Println("No se pudo recuperar la info del mbr.")
		return mbr, err
	}
	//checks if the mbr has available partitions
	if mbr.Partitions[3].Start != 0 {
		return mbr, errors.New("particiones ya definidas. no puede crear más")
	}

	return mbr, nil
}

func readDsk(file *os.File) (Mbr, error) {
	mbr := Mbr{}
	//Reading mbr
	file.Seek(0, 0)
	//converting data from binary
	data := readNextBytes(file, int(binary.Size(mbr)))
	buffer := bytes.NewBuffer(data)
	//writing mbr data
	err := binary.Read(buffer, binary.BigEndian, &mbr)
	if err != nil {
		fmt.Println("binary.Read failed at hello.go->readDisk()")
		return mbr, err
	}
	return mbr, nil
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

//checks if a previous partition already exists
func existName(name [16]byte, mbr Mbr, file *os.File) bool {
	var exists bool
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Name == name {
			return true
		}
		actualEbr := Ebr{}
		if mbr.Partitions[i].Type == 'E' { //looking for logic partitions
			actualPosition := mbr.Partitions[i].Start
			for true {
				file.Seek(int64(actualPosition), 0)
				//tries to obtain a possibly ebr
				data := readNextBytes(file, int(EBRSIZE))
				buffer := bytes.NewBuffer(data)
				err := binary.Read(buffer, binary.BigEndian, &actualEbr)
				if err != nil { //couldn't obtain ebr
					fmt.Println("binary.Read failed", err)
					return false
				}
				//ebr obtained. checking if the name already exists
				if actualEbr.Name == name {
					return true
				}
				//-1 if the ebr doesn't have a next
				if actualEbr.Next == -1 { //means that the last ebr was founded
					break
				} else {
					actualPosition = uint64(actualEbr.Next) //where is the next ebr (in case that exists)
				}
			}
		}
	}
	return exists
}
