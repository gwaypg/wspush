# Instruction
A distributed websocket node 

# Build
```
git clone https://github.com/gwaypg/wspush

cd wspush
. env.sh
sup build all

Or

. env.sh
cd cmd/wsnode
go build # or sup build
```

# Run
```
cd cmd/wsnode
./wsnode
```

# Deployment

## origin
```
PRJ_ROOT=<wspush root path> ./$PRJ_ROOT/cmd/wsnode/wsnode
```

## supd
```
git clone https://github.com/gwaypg/supd
cd supd
. env.sh
cd cmd/supd
./publish.sh
./setup.sh install # uninstall: ./setup.sh clean 

cd wspush
. env.sh
cd cmd/wsnode
sup install
sup status
```

TODO: more deployments

