#include "bridge.h"

/* Go trampolines (see //export in callbacks.go). */
extern void goOnTagArrival(nfc_tag_info_t *info);
extern void goOnTagDeparture(void);
extern void goOnSnepClientArrival(void);
extern void goOnSnepClientDeparture(void);
extern void goOnSnepServerArrival(void);
extern void goOnSnepServerDeparture(void);
extern void goOnSnepMessageReceived(unsigned char *message, unsigned int length);
extern void goOnHCEActivated(unsigned char mode);
extern void goOnHCEDeactivated(void);
extern void goOnHCEDataReceived(unsigned char *data, unsigned int length);
extern void goOnHandoverRequest(unsigned char *msg, unsigned int length);
extern void goOnHandoverSelect(unsigned char *msg, unsigned int length);
extern void goOnLLCPClientArrival(void);
extern void goOnLLCPClientDeparture(void);
extern void goOnLLCPServerArrival(void);
extern void goOnLLCPServerDeparture(void);
extern void goOnLLCPMessageReceived(void);

/* C struct field "type" is a Go keyword; expose via helper. */
nfc_handover_bt_type_t nfcgo_bt_type(nfc_btoob_pairing_t *p) {
    return p->type;
}

static nfcTagCallback_t g_tag_cb;
static nfcSnepClientCallback_t g_snep_client_cb;
static nfcSnepServerCallback_t g_snep_server_cb;
static nfcHostCardEmulationCallback_t g_hce_cb;
static nfcHandoverCallback_t g_handover_cb;
static nfcSnepClientCallback_t g_llcp_client_cb;
static nfcllcpConnlessServerCallback_t g_llcp_server_cb;

void nfcgo_register_tag_cb(void) {
    g_tag_cb.onTagArrival = goOnTagArrival;
    g_tag_cb.onTagDeparture = goOnTagDeparture;
    registerTagCallback(&g_tag_cb);
}

void nfcgo_deregister_tag_cb(void) {
    deregisterTagCallback();
}

int nfcgo_register_snep_client_cb(void) {
    g_snep_client_cb.onDeviceArrival = goOnSnepClientArrival;
    g_snep_client_cb.onDeviceDeparture = goOnSnepClientDeparture;
    return nfcSnep_registerClientCallback(&g_snep_client_cb);
}

void nfcgo_deregister_snep_client_cb(void) {
    nfcSnep_deregisterClientCallback();
}

int nfcgo_register_snep_server_cb(void) {
    g_snep_server_cb.onDeviceArrival = goOnSnepServerArrival;
    g_snep_server_cb.onDeviceDeparture = goOnSnepServerDeparture;
    g_snep_server_cb.onMessageReceived = goOnSnepMessageReceived;
    return nfcSnep_startServer(&g_snep_server_cb);
}

void nfcgo_stop_snep_server(void) {
    nfcSnep_stopServer();
}

void nfcgo_register_hce_cb(void) {
    g_hce_cb.onHostCardEmulationActivated = goOnHCEActivated;
    g_hce_cb.onHostCardEmulationDeactivated = goOnHCEDeactivated;
    g_hce_cb.onDataReceived = goOnHCEDataReceived;
    nfcHce_registerHceCallback(&g_hce_cb);
}

void nfcgo_deregister_hce_cb(void) {
    nfcHce_deregisterHceCallback();
}

int nfcgo_register_handover_cb(void) {
    g_handover_cb.onHandoverRequestReceived = goOnHandoverRequest;
    g_handover_cb.onHandoverSelectReceived = goOnHandoverSelect;
    return nfcHo_registerCallback(&g_handover_cb);
}

void nfcgo_deregister_handover_cb(void) {
    nfcHo_deregisterCallback();
}

int nfcgo_register_llcp_client_cb(void) {
    g_llcp_client_cb.onDeviceArrival = goOnLLCPClientArrival;
    g_llcp_client_cb.onDeviceDeparture = goOnLLCPClientDeparture;
    return nfcLlcp_ConnLessRegisterClientCallback(&g_llcp_client_cb);
}

void nfcgo_deregister_llcp_client_cb(void) {
    nfcLlcp_ConnLessDeregisterClientCallback();
}

int nfcgo_register_llcp_server_cb(void) {
    g_llcp_server_cb.onDeviceArrival = goOnLLCPServerArrival;
    g_llcp_server_cb.onDeviceDeparture = goOnLLCPServerDeparture;
    g_llcp_server_cb.onMessageReceived = goOnLLCPMessageReceived;
    return nfcLlcp_ConnLessStartServer(&g_llcp_server_cb);
}

void nfcgo_stop_llcp_server(void) {
    nfcLlcp_ConnLessStopServer();
}
