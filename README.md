# pi-psu

This is just a simple server that listens for messages from a Span device and power cycles the device if it hasn't seen any traffic from the device within some time span.

## Deployment

This runs on the host `relay` on the local network.  You can inspect the state at <http://relay:8080/> and trigger a reboot of the connected device at <http://relay:8080/reset>

## Example

Assuming you have set `PI_PSU_TOKEN` to your SPAN API token, and you are using the board Bjørn made with the relay attached to pin 27, the defaults for the other parameters should do.

```shell
pi-psu --collection 566ecm37cmcahj --device 566ecm37cmcahh
```

## Command line options

```shell
$ pi-psu -h
Usage:
  pi-psu [OPTIONS]

Application Options:
      --addr=                    HTTP interface listen address (default: :8080)
      --token=                   Span API token [$PI_PSU_TOKEN]
      --collection=              Span collection id [$PI_PSU_COLLECTION]
      --device=                  Span device ID [$PI_PSU_DEVICE]
      --msg-timeout=             time after last message seen from device to when we power cycle (default: 2m)
      --minimum-reboot-interval= minimum time between reboots (default: 5m)
      --gpio-pin=                GPIO pin for relay (default: 27)
      --gpio-hold=               how long we turn off power (default: 3s)

Help Options:
  -h, --help                     Show this help message

```
