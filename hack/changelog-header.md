### Linux

```shell
curl -L https://github.com/jenkins-x/jx-kcheck/releases/download/v{{.Version}}/jx-kcheck-linux-amd64.tar.gz | tar xzv 
sudo mv jx-kcheck /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-kcheck/releases/download/v{{.Version}}/jx-kcheck-darwin-amd64.tar.gz | tar xzv
sudo mv jx-kcheck /usr/local/bin
```

