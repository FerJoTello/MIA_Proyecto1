package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	Lexical "./elements"
)

var tokens []Lexical.Token    //tokens list from the command analyzed
var actualToken Lexical.Token //actual token that is being analyzed
var controlIndex int          //the index of the actual token

//EBRSIZE total size of the Ebr
const EBRSIZE = unsafe.Sizeof(ebr{})

//returns the totalSize of a ebr
func totalEbrSize(sizeOnBytes uint64) uint64 {
	return uint64(EBRSIZE) + sizeOnBytes
}

type partition struct {
	Name   [20]byte
	Type   byte
	Fit    byte
	Status byte
	Start  uint64
	Size   uint64
}

type mbr struct {
	Size          uint64
	CreationDate  [20]byte
	DiskSignature uint64
	Partitions    [4]partition
}

type ebr struct {
	Status byte
	Fit    byte
	Start  uint64
	Size   uint64
	Next   int64
	Name   [16]byte
}

func readDsk(file *os.File) (mbr, error) {
	mbr := mbr{}
	//Reading mbr
	file.Seek(0, 0)
	//converting data from binary
	data := readNextBytes(file, int(unsafe.Sizeof(mbr)))
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

func main() {
	main001()
}

func main001() {
	scanner := bufio.NewScanner(os.Stdin)
	var command string
	for command != "fin" {
		fmt.Print("Ingresa un comando:\n")
		scanner.Scan()
		command = "exec -pAth->\"/home/fernando/Documentos/2020 2do Semestre/Archivos/Proyecto1/pruebas/Comandos.mia\"" //scanner.Text()
		tokens = Lexical.Analyze(command)
		//is necessary to check if the tokens correspond to a command
		checkCommands()
		break
	}
}

//to check commands
func checkCommands() {
	fmt.Println("start checking commands...")
	controlIndex = 0
	actualToken = tokens[controlIndex] //esto se puede optimizar xd
	for controlIndex < len(tokens) {
		if compareActualToken(Lexical.ResExec) {
			nextToken()
			exec()
			break
		} else if compareActualToken(Lexical.ResMkdisk) {
			nextToken()
			mkdisk()
		} else if compareActualToken(Lexical.ResRmdisk) {
			nextToken()
			rmdisk()
		} else if compareActualToken(Lexical.ResFdisk) {
			nextToken()
			fdisk()
		}
		//a command has executed
		nextToken()
		if controlIndex == -1 {
			fmt.Println("Ya estufas")
			break
		}
	}
}

//fdisk
func fdisk() {
	var add, size int // variables used to hold the command values
	var path, name, unit, typee, fit, delete string
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResSize) {
				nextToken()
				size = getNumber()
			} else if compareActualToken(Lexical.ResPath) {
				nextToken()
				path = getPathString()
			} else if compareActualToken(Lexical.ResName) {
				nextToken()
				name = getID()
			} else if compareActualToken(Lexical.ResUnit) {
				nextToken()
				unit = getID()
			} else if compareActualToken(Lexical.ResType) {
				nextToken()
				typee = getID()
			} else if compareActualToken(Lexical.ResFit) {
				nextToken()
				fit = getID()
			} else if compareActualToken(Lexical.ResDelete) {
				nextToken()
				delete = getID()
			}
		} else {
			nextToken()
		}
	}
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
				createExtendedPartition(file, name, sizeOnBytes, fitOnByte, typeOnByte)
			} else if strings.ToLower(typee) == "l" { //logic partition
				typeOnByte = 'L'
				createLogicPartition(file, name, fitOnByte, sizeOnBytes)
			} else if strings.ToLower(typee) == "p" || typee == "" { //primary partition
				typeOnByte = 'P'
				createPrimaryPartition(file, name, sizeOnBytes, fitOnByte, typeOnByte)
			} else {
				fmt.Println("No se ha podido crear la particion. El tipo indicado no es valido.")
				return
			}
		}
	}
}

func createLogicPartition(file *os.File, partName string, fit byte, size uint64) {
	mbr, err := canCreatePartition(file)
	if err.Error() != "particiones ya definidas. no puede crear más" {
		fmt.Println("No se puede crear particion", err.Error())
		return
	}
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
	actualEbr := ebr{}
	ebrSize := int(unsafe.Sizeof(actualEbr))
	var ebrList []ebr //contains founded ebr
	//there could exist more than one ebr partition on the extended so it's necessary to find where its logic ends
	for true {
		file.Seek(int64(actualPosition), 0)
		//tries to obtain a possibly ebr
		data := readNextBytes(file, ebrSize)
		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &actualEbr)
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
	copy(actualEbr.Name[:], partName)
	//actualEbr.Start is not defined yet since it's necessary to find its position
	if len(ebrList) > 1 { //more than one ebr
		for i := 0; i < len(ebrList); i++ { //looking through the ebrs to find the first fit
			if ebrList[i].Next == -1 { //the next position is available (ebrList[i] is the last ebr)
				if totalEbrSize(actualEbr.Size) < (extended.Start+extended.Size)-(ebrList[i].Start+totalEbrSize(ebrList[i].Size)) { //ebr size should be less than the free space (total partition space - occupied space by the last ebr)
					actualEbr.Start = ebrList[i].Start + totalEbrSize(ebrList[i].Size)
					ebrList[i].Next = int64(actualEbr.Start)
					ebrList = append(ebrList, actualEbr)
					break
				}
			} else if ebrList[i].Next >= 0 { //next position is occupied
				if totalEbrSize(actualEbr.Size) < (ebrList[i+1].Start+totalEbrSize(ebrList[i+1].Size))-(ebrList[i].Start+totalEbrSize(ebrList[i].Size)) { //checks if there is enough space to set the actualEbr in between another two ebrs
					actualEbr.Start = totalEbrSize(ebrList[i].Size) + ebrList[i].Start
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
			if totalEbrSize(actualEbr.Size) < extended.Size { //ebr size should be less than extended partition total size
				actualEbr.Start = extended.Start
				ebrList[0] = actualEbr
			} else {
				fmt.Println("El tamano del nuevo ebr es mayor que la partición que lo contiene")
				return
			}
		} else { //First ebr is not empty
			if totalEbrSize(actualEbr.Size) < extended.Size-totalEbrSize(ebrList[0].Size) { //ebr size should be less than extended partition free space
				ebrList[0].Next = int64(ebrList[0].Start + totalEbrSize(ebrList[0].Size))
				actualEbr.Start = ebrList[0].Start + totalEbrSize(ebrList[0].Size)
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

func createExtendedPartition(file *os.File, partName string, size uint64, fit byte, typee byte) {
	mbr, err := canCreatePartition(file)
	if err != nil {
		fmt.Println("No se puede crear particion", err.Error())
		return
	}
	//creates a new partition with the provided name, size and the worst adjustment
	newPartition := partition{}
	copy(newPartition.Name[:], partName)
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
	ebr := ebr{Start: newPartition.Start, Next: -1}
	file.Seek(int64(ebr.Start), 0)
	binario.Reset()
	binary.Write(&binario, binary.BigEndian, &ebr)
	escribirBytes(file, binario.Bytes())
}

func canCreatePartition(file *os.File) (mbr, error) {
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

func createPrimaryPartition(file *os.File, partName string, size uint64, fit byte, typee byte) {
	//checking if partitions are available
	mbr, err := canCreatePartition(file)
	if err != nil {
		fmt.Println("No se puede crear particion", err.Error())
		return
	}
	//creates a new partition with the provided name, size and adjustment
	newPartition := partition{}
	copy(newPartition.Name[:], partName)
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

//obtains the new partition start position and the mbr partitions' index
func getStartPartition(size uint64, mbr mbr) (uint64, int) {
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
		if size < mbr.Size-uint64(unsafe.Sizeof(mbr)) { //size of the new partition should be less than the free space
			return uint64(unsafe.Sizeof(mbr)), 0
		}
	}
	return 0, -1
}

func orderPartitionsByStart(m mbr) mbr {
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

//rmdisk
func rmdisk() {
	var path string
	if compareActualToken(Lexical.SMinus) {
		nextToken()
		//next token should be path since rmdisk just has one param
		if compareActualToken(Lexical.ResPath) {
			nextToken()
			path = getPathString()
		}
	}
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
	fmt.Println("No se ha podido eliminar el disco.")
}

//to check commands
func compareActualToken(comparedType Lexical.TokenType) bool {
	return actualToken.Type == comparedType
}

//to check commands
func nextToken() {
	controlIndex++
	if controlIndex < len(tokens) {
		actualToken = tokens[controlIndex] //no problem
	} else {
		controlIndex = -1 //there are no more elements on the array
	}
}

//command exec
//exec -path->/valid/file.mia
func exec() {
	var path string
	//exec was detected. the next token should be "-"
	if compareActualToken(Lexical.SMinus) {
		nextToken()
		//next token should be path since exec just has one param
		if compareActualToken(Lexical.ResPath) {
			nextToken()
			path = getPathString()
		}
	}
	if path != "" {
		content := readFile(path)
		if content != "" {
			tokens = Lexical.Analyze(content)
			checkCommands()
		}
	}
}

func readFile(path string) string {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		return string(content)
	}
	fmt.Println("Ocurrió un error al leer el archivo:\n" + err.Error())
	return ""
}

//command mkdisk
func mkdisk() {
	var size int // variables used to hold the command values
	var path, name, unit string
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResSize) {
				nextToken()
				size = getNumber()
			} else if compareActualToken(Lexical.ResPath) {
				nextToken()
				path = getPathString()
			} else if compareActualToken(Lexical.ResName) {
				nextToken()
				name = getFileName()
			} else if compareActualToken(Lexical.ResUnit) {
				nextToken()
				unit = getID()
			}
		} else {
			nextToken()
		}
	}
	//once the parameters are defined the disk can be created
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
		var partitions [4]partition
		for i := 0; i < 4; i++ {
			copy(partitions[i].Name[:], "new_part")
			partitions[i].Status = 'F'
			partitions[i].Fit = 'x'
			partitions[i].Type = 'x'
			partitions[i].Start = 0
			partitions[i].Size = 0
		}
		// disk init
		newMbr := mbr{}
		newMbr.Size = sizeOnBytes
		newMbr.Partitions = partitions
		newMbr.DiskSignature = uint64(time.Now().Unix())
		copy(newMbr.CreationDate[:], time.Now().String()[0:19])
		writeMbr(path, name, sizeOnBytes, newMbr)
	} else {
		fmt.Println("No es posible ejecutar el comando \"mkdisk\". Parametros no definidos")
	}

}

func writeMbr(path string, fileName string, sizeOnBytes uint64, newMbr mbr) {
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
	fmt.Println("File created succesfully xd")
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

//params
func getNumber() int {
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.Number) {
			value, err := strconv.Atoi(actualToken.Value)
			if err == nil {
				nextToken()
				return value
			}
		}
	}
	return -1
}

func getPathString() string {
	var path string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.Path) {
			path = actualToken.Value
			nextToken()
		} else if compareActualToken(Lexical.String) {
			path = strings.Trim(actualToken.Value, "\"")
			nextToken()
		}
	}
	return path
}

func getFileName() string {
	var name string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.FileName) {
			name = actualToken.Value
			nextToken()
		}
	}
	return name
}

func getID() string {
	var unit string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.ID) {
			unit = actualToken.Value
			nextToken()
		}
	}
	return unit
}
