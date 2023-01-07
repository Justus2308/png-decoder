#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR/src
go build -a main.go
echo "compiled executable"
cd $DIR
cat > run.sh <<EOF
#!/bin/bash

$DIR/src/main
EOF
echo "created run.sh"
chmod +x run.sh