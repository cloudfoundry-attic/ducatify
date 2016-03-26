# ducatify
sprinkle some ducati on that diego

```bash
go get github.com/cloudfoundry-incubator/ducatify/cmd/ducatify

ducatify \
   -diego path/to/my/diego-deployment-manifest.yml \
   > diego-with-ducati.yml

bosh -d diego-with-ducati.yml deploy
```
