# m3u8 list parser and channel list generator
## Requirements:
* Install gonalg
* Install postgres
* Install golang task (https://taskfile.dev/installation/)
* Install golang migrate (https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)
* Rename m3u8.env.example to m3u8.env and set own config variables
* Rename order.yaml.example to order.yaml and make own output formatter list
* Run cmd: task migrate
* Ready to run, first channel parsing iteration could take some time for DB fill up with channel resolution
