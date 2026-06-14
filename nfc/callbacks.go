package nfc

/*
#include "bridge.h"
*/
import "C"

import (
	"sync"
	"unsafe"
)

var (
	tagMu sync.RWMutex
	tagH  TagHandler

	snepClientMu sync.RWMutex
	snepClientH  PeerHandler

	snepServerMu sync.RWMutex
	snepServerH  SnepServerHandler

	hceMu sync.RWMutex
	hceH  HCEHandler

	handoverMu sync.RWMutex
	handoverH  HandoverHandler

	llcpClientMu sync.RWMutex
	llcpClientH  PeerHandler

	llcpServerMu sync.RWMutex
	llcpServerH  LLCPHandler
)

func setTagHandler(h TagHandler) {
	tagMu.Lock()
	tagH = h
	tagMu.Unlock()
}

func clearTagHandler() {
	setTagHandler(TagHandler{})
}

func setSnepClientHandler(h PeerHandler) {
	snepClientMu.Lock()
	snepClientH = h
	snepClientMu.Unlock()
}

func clearSnepClientHandler() {
	setSnepClientHandler(PeerHandler{})
}

func setSnepServerHandler(h SnepServerHandler) {
	snepServerMu.Lock()
	snepServerH = h
	snepServerMu.Unlock()
}

func clearSnepServerHandler() {
	setSnepServerHandler(SnepServerHandler{})
}

func setHCEHandler(h HCEHandler) {
	hceMu.Lock()
	hceH = h
	hceMu.Unlock()
}

func clearHCEHandler() {
	setHCEHandler(HCEHandler{})
}

func setHandoverHandler(h HandoverHandler) {
	handoverMu.Lock()
	handoverH = h
	handoverMu.Unlock()
}

func clearHandoverHandler() {
	setHandoverHandler(HandoverHandler{})
}

func setLLCPClientHandler(h PeerHandler) {
	llcpClientMu.Lock()
	llcpClientH = h
	llcpClientMu.Unlock()
}

func clearLLCPClientHandler() {
	setLLCPClientHandler(PeerHandler{})
}

func setLLCPHandler(h LLCPHandler) {
	llcpServerMu.Lock()
	llcpServerH = h
	llcpServerMu.Unlock()
}

func clearLLCPHandler() {
	setLLCPHandler(LLCPHandler{})
}

//export goOnTagArrival
func goOnTagArrival(info *C.nfc_tag_info_t) {
	tagMu.RLock()
	fn := tagH.OnDiscovered
	tagMu.RUnlock()
	if fn != nil {
		fn(tagFromC(info))
	}
}

//export goOnTagDeparture
func goOnTagDeparture() {
	tagMu.RLock()
	fn := tagH.OnRemoved
	tagMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnSnepClientArrival
func goOnSnepClientArrival() {
	snepClientMu.RLock()
	fn := snepClientH.OnConnected
	snepClientMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnSnepClientDeparture
func goOnSnepClientDeparture() {
	snepClientMu.RLock()
	fn := snepClientH.OnDisconnected
	snepClientMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnSnepServerArrival
func goOnSnepServerArrival() {
	snepServerMu.RLock()
	fn := snepServerH.OnPeerConnected
	snepServerMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnSnepServerDeparture
func goOnSnepServerDeparture() {
	snepServerMu.RLock()
	fn := snepServerH.OnPeerDisconnected
	snepServerMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnSnepMessageReceived
func goOnSnepMessageReceived(message *C.uchar, length C.uint) {
	snepServerMu.RLock()
	fn := snepServerH.OnMessage
	snepServerMu.RUnlock()
	if fn != nil && message != nil && length > 0 {
		fn(C.GoBytes(unsafe.Pointer(message), C.int(length)))
	}
}

//export goOnHCEActivated
func goOnHCEActivated(mode C.uchar) {
	hceMu.RLock()
	fn := hceH.OnActivated
	hceMu.RUnlock()
	if fn != nil {
		fn(uint8(mode))
	}
}

//export goOnHCEDeactivated
func goOnHCEDeactivated() {
	hceMu.RLock()
	fn := hceH.OnDeactivated
	hceMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnHCEDataReceived
func goOnHCEDataReceived(data *C.uchar, length C.uint) {
	hceMu.RLock()
	fn := hceH.OnAPDU
	hceMu.RUnlock()
	if fn != nil && data != nil && length > 0 {
		fn(C.GoBytes(unsafe.Pointer(data), C.int(length)))
	}
}

//export goOnHandoverRequest
func goOnHandoverRequest(msg *C.uchar, length C.uint) {
	handoverMu.RLock()
	fn := handoverH.OnRequest
	handoverMu.RUnlock()
	if fn != nil && msg != nil && length > 0 {
		fn(C.GoBytes(unsafe.Pointer(msg), C.int(length)))
	}
}

//export goOnHandoverSelect
func goOnHandoverSelect(msg *C.uchar, length C.uint) {
	handoverMu.RLock()
	fn := handoverH.OnSelect
	handoverMu.RUnlock()
	if fn != nil && msg != nil && length > 0 {
		fn(C.GoBytes(unsafe.Pointer(msg), C.int(length)))
	}
}

//export goOnLLCPClientArrival
func goOnLLCPClientArrival() {
	llcpClientMu.RLock()
	fn := llcpClientH.OnConnected
	llcpClientMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnLLCPClientDeparture
func goOnLLCPClientDeparture() {
	llcpClientMu.RLock()
	fn := llcpClientH.OnDisconnected
	llcpClientMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnLLCPServerArrival
func goOnLLCPServerArrival() {
	llcpServerMu.RLock()
	fn := llcpServerH.OnPeerConnected
	llcpServerMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnLLCPServerDeparture
func goOnLLCPServerDeparture() {
	llcpServerMu.RLock()
	fn := llcpServerH.OnPeerDisconnected
	llcpServerMu.RUnlock()
	if fn != nil {
		fn()
	}
}

//export goOnLLCPMessageReceived
func goOnLLCPMessageReceived() {
	llcpServerMu.RLock()
	fn := llcpServerH.OnMessage
	llcpServerMu.RUnlock()
	if fn != nil {
		fn()
	}
}
