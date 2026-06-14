package nfc

// Target types (TARGET_TYPE_*).
const (
	TargetUnknown       = -1
	TargetISO14443_3A   = 1
	TargetISO14443_3B   = 2
	TargetISO14443_4    = 3
	TargetFelica        = 4
	TargetISO15693      = 5
	TargetNDEF          = 6
	TargetNDEFFormat    = 7
	TargetMifareClassic = 8
	TargetMifareUL      = 9
	TargetKovioBarcode  = 10
	TargetISO14443_3A3B = 11
)

// Listen modes (MODE_LISTEN_*).
const (
	ListenModeA = 0x80
	ListenModeB = 0x81
	ListenModeF = 0x82
)

// Technology masks (NFA_TECHNOLOGY_MASK_*).
const (
	TechDefault  = -1
	TechA        = 0x01
	TechB        = 0x02
	TechF        = 0x04
	TechISO15693 = 0x08
	TechKovio    = 0x20
	TechAActive  = 0x40
	TechFActive  = 0x80
	TechAll      = 0xFF
)

// NDEF type name format (NDEF_TNF_*).
const (
	TNFEmpty     = 0
	TNFWellKnown = 1
	TNFMedia     = 2
	TNFURI       = 3
	TNFExt       = 4
	TNFUnknown   = 5
	TNFUnchanged = 6
)

// Protocols (NFA_PROTOCOL_*).
const (
	ProtocolUnknown = 0x00
	ProtocolT1T     = 0x01
	ProtocolT2T     = 0x02
	ProtocolT3T     = 0x03
	ProtocolISODep  = 0x04
	Protocol15693   = 0x06
	ProtocolMifare  = 0x80
)

// HCE flags.
const (
	HCESkipNDEFCheck = 0x80
	HCEEnabled       = 0x01
)

// RecordType is the decoded high-level NDEF record type.
type RecordType int

const (
	RecordText  RecordType = 0
	RecordURI   RecordType = 1
	RecordHS    RecordType = 2
	RecordHR    RecordType = 3
	RecordOther RecordType = 4
)

// BluetoothCarrierType is the handover bluetooth carrier type.
type BluetoothCarrierType int

const (
	BluetoothCarrierUnknown BluetoothCarrierType = 0
	BluetoothCarrierClassic BluetoothCarrierType = 1
	BluetoothCarrierBLE     BluetoothCarrierType = 2
)

// CarrierPowerState is the handover carrier power state.
type CarrierPowerState int

const (
	CarrierInactive   CarrierPowerState = 0
	CarrierActive     CarrierPowerState = 1
	CarrierActivating CarrierPowerState = 2
	CarrierUnknown    CarrierPowerState = 3
)

// Tag describes a detected NFC tag.
type Tag struct {
	Technology uint
	Handle     uint
	UID        []byte
	Protocol   uint8
}

// NDEFInfo describes NDEF capabilities of a tag.
type NDEFInfo struct {
	Length    uint
	MaxLength uint
	Writable  bool
}

// BluetoothPairing holds bluetooth handover pairing data.
type BluetoothPairing struct {
	PowerState CarrierPowerState
	Type       BluetoothCarrierType
	NDEF       []byte
	Address    [6]byte
	DeviceName []byte
}

// BluetoothRequest holds bluetooth handover request data.
type BluetoothRequest = BluetoothPairing

// WiFiPairing holds wifi handover select data.
type WiFiPairing struct {
	PowerState CarrierPowerState
	NDEF       []byte
	SSID       []byte
	Key        []byte
}

// WiFiRequest holds wifi handover request data.
type WiFiRequest struct {
	HasWiFi bool
	NDEF    []byte
}

// HandoverRequest holds parsed handover request info.
type HandoverRequest struct {
	Bluetooth BluetoothRequest
	WiFi      WiFiRequest
}

// HandoverSelect holds parsed handover select info.
type HandoverSelect struct {
	Bluetooth BluetoothPairing
	WiFi      WiFiPairing
}

// TagHandler is called when a tag enters or leaves the field.
type TagHandler struct {
	OnDiscovered func(Tag)
	OnRemoved    func()
}

// SnepServerHandler is called for SNEP server events.
type SnepServerHandler struct {
	OnPeerConnected    func()
	OnPeerDisconnected func()
	OnMessage          func(message []byte)
}

// PeerHandler is called when a peer device is detected or removed.
type PeerHandler struct {
	OnConnected    func()
	OnDisconnected func()
}

// LLCPHandler is called for connectionless LLCP server events.
type LLCPHandler struct {
	OnPeerConnected    func()
	OnPeerDisconnected func()
	OnMessage          func()
}

// HCEHandler is called for host card emulation events.
type HCEHandler struct {
	OnActivated   func(mode uint8)
	OnDeactivated func()
	OnAPDU        func(data []byte)
}

// HandoverHandler is called for handover events.
type HandoverHandler struct {
	OnRequest func(message []byte)
	OnSelect  func(message []byte)
}

// DiscoveryOptions configures RF discovery.
type DiscoveryOptions struct {
	Technologies int  // bitmask of Tech* constants; TechAll for everything
	ReaderOnly   bool // disable P2P and card emulation
	HostRouting  bool // enable host card emulation routing
	Restart      bool // force restart discovery
}

// DefaultDiscovery matches nfcDemoApp poll mode (DEFAULT_NFA_TECH_MASK, no restart).
func DefaultDiscovery() DiscoveryOptions {
	return DiscoveryOptions{
		Technologies: TechDefault,
		ReaderOnly:   false,
		Restart:      false,
	}
}
