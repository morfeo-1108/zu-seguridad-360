# Vault en modo servidor (no dev): los secretos persisten en el volumen vault_data
# y Vault arranca sellado tras cada reinicio (ver scripts/vault-unseal.sh).
storage "file" {
  path = "/vault/file"
}

# Solo accesible dentro de la red Docker zu-network y en 127.0.0.1 del host.
# TLS deshabilitado en este tramo interno: ver docs/INFRAESTRUCTURA.md (pendiente).
listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = true
}

api_addr      = "http://vault:8200"
disable_mlock = false
ui            = false
