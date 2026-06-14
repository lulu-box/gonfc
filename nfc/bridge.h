#ifndef NFCGO_BRIDGE_H
#define NFCGO_BRIDGE_H

#include "linux_nfc_api.h"

void nfcgo_register_tag_cb(void);
void nfcgo_deregister_tag_cb(void);
int nfcgo_register_snep_client_cb(void);
void nfcgo_deregister_snep_client_cb(void);
int nfcgo_register_snep_server_cb(void);
void nfcgo_stop_snep_server(void);
void nfcgo_register_hce_cb(void);
void nfcgo_deregister_hce_cb(void);
int nfcgo_register_handover_cb(void);
void nfcgo_deregister_handover_cb(void);
int nfcgo_register_llcp_client_cb(void);
void nfcgo_deregister_llcp_client_cb(void);
int nfcgo_register_llcp_server_cb(void);
void nfcgo_stop_llcp_server(void);

nfc_handover_bt_type_t nfcgo_bt_type(nfc_btoob_pairing_t *p);

#endif
