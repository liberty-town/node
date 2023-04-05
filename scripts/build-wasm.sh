buildFlag="pandora-pay/config.BUILD_VERSION"
frontend="../frontend/"
mainWasmOutput="LibertyTown-main.wasm"

if [ $# -eq 0 ]; then
  echo "arguments missing"
fi

if [[ "$*" == "help" ]]; then
    echo "main, test|dev|build, brotli|zopfli|gzip"
    exit 1
fi

gitVersion=$(git log -n1 --format=format:"%H")
gitVersionShort=${gitVersion:0:12}

src=""
buildOutput="./dist/"

if [[ "$*" == *test* ]]; then
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "${buildOutput}/wasm_exec.js"
fi

if [[ "$*" == *main* ]]; then
  buildOutput+="main"
  src="./builds/webassembly/"
else
  echo "argument main missing"
  exit 1
fi

if [[ "$*" == *test* ]]; then
  buildOutput+="-test"
elif [[ "$*" == *dev* ]]; then
  buildOutput+="-dev"
elif [[ "$*" == *build* ]]; then
  buildOutput+="-build"
else
  echo "argument test|dev|build missing"
  exit 1
fi

buildOutput+=".wasm"

go version
(cd ${src} && GOOS=js GOARCH=wasm go build  -ldflags="-s -w" -o ${buildOutput} )
#( cd ${src} && tinygo build -o ${buildOutput}  -target wasm -size full  -no-debug -gc=leaking )

buildOutput=${src}${buildOutput}

finalOutput=${frontend}

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "${finalOutput}src/webworkers/dist/wasm_exec.js"
#cp "/usr/local/lib/tinygo/targets/wasm_exec.js" "${finalOutput}src/webworkers/dist/wasm_exec.js"

finalOutput+="dist/"

mkdir -p "${finalOutput}"

if [[ "$*" == *dev* ]]; then
  finalOutput+="dev/"
elif [[ "$*" == *build* ]]; then
  finalOutput+="build/"
fi

if ! [[ "$*" == *test* ]]; then

  mkdir -p "${finalOutput}"
  mkdir -p "${finalOutput}wasm"

  stat --printf="%s \n" ${buildOutput}

  echo "Deleting..."

  rm ${buildOutput}.br 2>/dev/null
  rm ${buildOutput}.gz 2>/dev/null

  if [[ "$*" == *main* ]]; then
    finalOutput+="wasm/${mainWasmOutput}"
  fi

  echo "Copy to frontend/dist..."
  cp ${buildOutput} ${finalOutput}
fi

if [[ "$*" == *build* ]]; then

  if [[ "$*" == *brotli* ]]; then
    echo "Zipping using brotli..."
    if ! brotli -o ${buildOutput}.br ${buildOutput}; then
      echo "sudo apt-get install brotli"
      exit 1
    fi
    stat --printf="brotli size %s \n" ${buildOutput}.br
    echo "Copy to frontend/dist..."
    cp ${buildOutput}.br ${finalOutput}.br
  fi

  if [[ "$*" == *zopfli* ]]; then
    echo "Zipping using zopfli..."
    if ! zopfli ${buildOutput}; then
      echo "sudo apt-get install zopfli"
      exit 1
    fi
    stat --printf="zopfli gzip size: %s \n" ${buildOutput}.gz
    echo "Copy to frontend/build..."
    cp ${buildOutput}.gz ${finalOutput}.gz
  elif [[ "$*" == *gzip* ]]; then
    echo "Gzipping..."
    gzip --best ${buildOutput}
    stat --printf="gzip size %s \n" ${buildOutput}.gz
    echo "Copy to frontend/build..."
    cp ${buildOutput}.gz ${finalOutput}.gz
  fi

fi