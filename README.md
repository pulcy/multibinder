# Multibinder go client

This library is used to connect to Github's [multibinder](https://github.com/github/multibinder).

## Sample usage 

### Step 1: Build multibinder docker image 

```
cd multibinder 
docker build -t multibinder .
```

### Step 2: Build example app & docker image 

```
cd example 
./build.sh
```

### Step 3: Run multibinder 

```
docker create volume multibinder 
docker run --net=host -v multibinder:/opt/multibinder multibinder 
```

### Step 4: Run example

```
docker run -v multibinder:/opt/multibinder mbexample -socket=/opt/multibinder/multibinder.sock 
curl http://<dockerip>:8083
```

Then to really test it:

1. Kill the docker container running the example 
2. Start another curl command (it will wait for a connection that is not yet handled)
3. Restart the docker mbexample test and see the result appear from curl.