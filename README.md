yaproxy
=======
[![Build Status](https://travis-ci.org/denghongcai/yaproxy.svg?branch=master)](https://travis-ci.org/denghongcai/yaproxy)

# Install
download latest release file from [GitHub Release](https://github.com/denghongcai/yaproxy/releases) and choose [gfwlist.txt](https://raw.githubusercontent.com/denghongcai/yaproxy/master/gfwlist.txt) or [proxy.pac](https://raw.githubusercontent.com/denghongcai/yaproxy/master/proxy.pac) or both to download.

# Usage
type

```
yaproxy -h
```

to get help

# ss-config Example

### client side

```
{
    "server":"www.dhc.house",
    "server_port":443,
    "local_port":1080,
    "password":"boom",
    "method": "rc4-md5",
    "timeout":600,
    "suft": {
        "bandwidth": 25
    }
}
```

### client side

```
{
    "server":"0.0.0.0",
    "server_port":443,
    "local_port":1080,
    "password":"boom",
    "method": "rc4-md5",
    "timeout":600,
    "suft": {
        "local_addr": ":443",
        "bandwidth": 25
    }
}
```
