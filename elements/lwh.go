package elements

type superBoot struct {
	nameHD                       [16]byte //nombre del disco virtual
	countdirectoryTree           uint32   //cantidad de estructuras en el arbol de directorio
	countDirectoryDetail         uint32   //cantidad de estructuras en el detalle de directorio
	countiNodes                  uint32   //cantidad de inodos en la tabla de inodos
	countDataBlocks              uint32   //cantidad de bloque de datos libres
	freedirectoryTree            uint32   //cantidad de estructuras en el arbol de directorio libres
	freeDirectoryDetail          uint32   //cantidad de estructuras en el detalle de directorio libres
	freeiNodes                   uint32   //cantidad de inodos en la tabla de inodos libres
	freeDataBlocks               uint32   //cantidad de bloque de datos
	creationDate                 [20]byte //fecha de creacion tiene que ser formato: dd/mm/yyyy hh:mm
	lastMountDate                [20]byte //mismo formato que creationDate
	directoryTreeBitMapPointer   uint64   //apuntador al inicio del bitmap del arbol virtual
	directoryTreePointer         uint64   //apuntador al inicio del arbol virtual de directorio
	directoryDetailBitMapPointer uint64   //apuntador al inicio del bitmap del detalle directorio
	directoryDetailTreePointer   uint64   //apuntador al inicio del detalle directorio
	iNodeTableBitMapPointer      uint64   //apuntador al inicio del bitmap de la tabla de inodos
	iNodeTablePointer            uint64   //apuntador al inicio de la tabla de inodos
	dataBlockBitMapPointer       uint64   //apuntador al inicio del bitmap del bloque de datos
	dataBlockPointer             uint64   //apuntador al inicio del bloque de datos
	logPointer                   uint64   //apuntador al inicio del log o bitacora
	directoryTreeStructSize      uint32   //tamano de la estructura del arbol virtual de directorio
	directoryDetailStructSize    uint32   //tamano de la estructura del detalle de directorio
	iNodeStructSize              uint32   //tamano de la estructura del inodo
	dataBlockStructSize          uint32   //tamano de la estructura del bloque de datos
	directoryTreeFirstFreeBit    uint64   //primer bit libre del arbol de directorio
	directoryDetailFirstFreeBit  uint64   //primer bit libre del detalle directorio
	iNodeTableFirstFreeBit       uint64   //primer bit libre de la tabla de inodos
	dataBlockFirstFreeBit        uint64   //primer bit libre del bloque de datos
	magicNum                     uint32   //carnet
}

//arbol virtual de directorio
type directoryTree struct {
	creationDate               [20]byte  //dd/mm/yyyy hh:mm
	directoryName              [16]byte  //nombre del directorio
	subDirectoriesPointerArray [6]uint64 //apuntadores directos a subdirectorios
	directoryDetailPointer     uint64    //apuntador a un detalle de directorio
	directoryTreePointer       uint64    //apuntador al mismo tipo de estructura si se usan los 6 para almacenar mas
	proper                     [16]byte  //id del propietario
}

//detalle de directorio
type directoryDetail struct {
	files                  [5]myFile //arreglo de la info de cada archivo
	directoryDetailPointer uint64    //puntero a otro detalle directorio para almacenar mas archivos
}

type myFile struct {
	name         [16]byte //nombre del archivo
	iNodePointer uint64   //apuntador al iNodo
	creationDate [20]byte //dd/mm//yyyy hh:mm
	changeDate   [20]byte //ultima modificacion
}

//tabla de iNodo
type iNode struct {
	count            uint32    //numero de iNodo
	sizeFile         uint64    //tamano del archivo
	asignedBlocks    uint32    //numero de bloques asignados
	dataBlockPointer [4]uint64 //apuntadores a bloques de datos
	indirectPointer  uint64    //apuntador indirecto si el archivo utiliza mas de 4
}

type dataBlock struct {
	data [25]byte //info del archivo
}

type myLog struct {
	operationType byte     //tipo de operacion a realizarse
	logType       byte     //archivo = 0, directorio = 1
	name          [16]byte //nombre de archivo o directorio
	content       [100]byte
	date          [20]byte
}
