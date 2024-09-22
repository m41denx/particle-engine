# Particle Engine

### Architectures:

| OS                       | code   |
|--------------------------|--------|
| Windows x64              | `w64`  |
| Linux x64                | `l64`  |
| Linux arm64              | `l64a` |
| macOS x64 10.14+         | `d64`  |
| macOS arm64 11+          | `d64a` |

`w=windows, l=linux, d=darwin`

## API

### Fetch Particle Manifest
> URL: `GET SERVER_URL/repo/:author/:name@version/:archb`
>
> EXAMPLE: `GET https://hub.fruitspace.one/repo/m41den/thebox@latest/w64`

### Push Particle Manifest
> URL: `POST uname:token@SERVER_URL/upload/:name@version/:archb`
> 
> EXAMPLE: `POST https://m41den:mytoken@hub.fruitspace.one/upload/thebox@latest/w64`
>
> BODY: `application/json {...manifest...}`

### Push Particle Layer
> URL: `POST uname:token@SERVER_URL/upload/:name@version/:archb`
> 
> EXAMPLE: `POST https://m41den:mytoken@hub.fruitspace.one/upload/thebox@latest/w64`
> 
> BODY: `form-data/file <binary>`

### Pull Particle Layer
> URL: `GET SERVER_URL/layers/:layerid`
>
> EXAMPLE: `GET https://hub.fruitspace.one/layers/babcae0de0`

### Auth user
> URL: `GET uname:token@SERVER_URL/user`
> 
> EXAMPLE: `GET https://m41den:mytoken@hub.fruitspace.one/user`


## Config Structure
Language: [YAML](https://gopkg.in/yaml.v3)

### Config Structure
```yaml
name: author/particle_name@v2

meta:
    author: "Particle Author"
    note: "Short note"

layer:
    block: "[sha256 autogen]"
    server: "http://optional/v1"

recipe:
    - use: fruitspace/gd_android@2.2
    - apply: https://someothermod.xyz/particles/fifuser/gd_patcher
      env:
        AMOGUS: YES
      command: overriden

runnable:
  runner: "full"  # runner: thin/full (full is cygwin, 65->250mb)
  require:
    - apply: m41den/gd_patcher
  build:
    - run: |
        command -arg value
    - copy: "$MOD/file1"
      to: "$BUILD/file2"
  expose:
    python3: "/usr/bin/python.exe"
```

### FS Layout
```
/build *
/dev
/etc
/home
/opt
/runnable *
     ---> /<runnablename>   
/tmp
     ---> /buildcache
/usr
/var
integrity.json
```

