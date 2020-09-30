### Linux

```shell
curl -L https://github.com/jenkins-x/jx-health/releases/download/v{{.Version}}/jx-health-linux-amd64.tar.gz | tar xzv 
sudo mv jx-health /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-health/releases/download/v{{.Version}}/jx-health-darwin-amd64.tar.gz | tar xzv
sudo mv jx-health /usr/local/bin
```

