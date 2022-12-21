//go:build !js
// +build !js

package arguments

var Text = `libertytown.

Usage:
  libertytown [--debug] [--network=network] [--instance=prefix] [--instance-id=id] [--node-consensus=type] [--display-identity] [--display-apps] [--tcp-server-port=PORT] [--tcp-server-address=ADDRESS] [--store-data-type=type] [--store-settings-type=TYPE] [--serve-federation=FED]  [--settings-import-secret-mnemonic=mnemonic] [--settings-import-secret-entropy=entropy] [--tcp-server-tls-cert-file=path] [--tcp-server-tls-key-file=path] [--tcp-max-clients=limit] [--tcp-max-server-sockets=limit] [--tcp-connections-ready=threshold] [--tcp-server-url=url]
  libertytown -v | --version

Options:
  -h --version                                        Show version.
  --debug                                             Debug flag set enabled.
  --network=network                                   Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet].
  --instance=prefix                                   Prefix of the instance [default: 0].
  --instance-id=id                                    Number of forked instance (when you open multiple instances). It should be a string number like "1","2","3","4" etc
  --node-consensus=type                               Consensus type. Accepted values: "full|app|none" [default: full].
  --display-identity                                  Display your identity.
  --display-apps                                      Display all federations and chats.
  --tcp-server-tls-cert-file=path                     Load TLS certificate file from given path.
  --tcp-server-tls-key-file=path                      Load TLS ke file from given path.
  --tcp-server-url=url                                TCP Server URL (schema, address, port, path).
  --tcp-server-port=PORT                              Server Port. [default: 8080].
  --tcp-server-address=ADDRESS                        Server Address.
  --tcp-max-clients=limit                             Change limit of clients [default: 50].
  --tcp-connections-ready=threshold                   Number of connections to become "ready" state [default: 1].
  --tcp-max-server-sockets=limit                      Change limit of servers [default: 500].
  --store-data-type=TYPE                              Storage method for Data. [default: bolt]
  --store-settings-type=TYPE                          Storage method for Settings. [default: bolt]
  --settings-import-secret-mnemonic=mnemonic          Import settings from a given Mnemonic. It will delete your existing settings. 
  --settings-import-secret-entropy=entropy            Import settings from a given Entropy. It will delete your existing settings.
  --serve-federation=FED                               List of Federations to distribute. 
`
