# udp2tcp

works fine now,^ ^ 

## sample

app-client(udpclient) → udp2tcp-client → udp2tcp-server → app-server(udpserver)

```sh udp2tcp-client
udp2tcp -c :11111 -a  (udp2tcp-server's ip):17002
``` 

```sh udp2tcp-server
udp2tcp -s :17002 -z  (app-server's ip):17002
```