# golinks
Simple url shortener that can be used at home. Redirects added to the url shortener are automatically saved into a json configuration file.

### Install
Download the code and build in golang for your host. Setup your local dns to point to the host. Typically setup something simple like "go" as the name of the host. 

### Run
```
./golinks --http_port=80 &
```

### Adding new links
```
http://go/add/<shortname>/<redirect url> 
```

### Deleting links
``` 
http://go/del/<shortname> 
```

### Show all redirects
```
http://go/list
```

### Using the redirector
```
http://go/shortname
```
