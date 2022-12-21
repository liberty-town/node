//go:build js
// +build js

package arguments

var Text = `libertytown.

Usage:
  libertytown [--debug] [--network=network] [--instance=prefix] [--instance-id=id] [--node-consensus=type] [--display-identity] [--display-apps] [--store-data-type=type] [--store-settings-type=TYPE] [--tcp-max-clients=limit] [--tcp-connections-ready=threshold] [--tcp-connect-onion-addresses] 
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
  --tcp-max-clients=limit                             Change limit of clients [default: 50].
  --tcp-connections-ready=threshold                   Number of connections to become "ready" state [default: 1].
  --tcp-connect-onion-addresses                       If it will connect to onion addresses.
  --store-data-type=TYPE                              Storage method for Data. [default: js]
  --store-settings-type=TYPE                          Storage method for Settings. [default: js]
`
