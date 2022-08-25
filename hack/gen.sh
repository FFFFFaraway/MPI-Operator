mkdir -p ./api/batch.test.bdap.com
cp -r ./api/v1 ./api/batch.test.bdap.com
./hack/generate-groups.sh "client,informer,lister" \
  ./client \
  ./api \
  batch.test.bdap.com:v1 \
  --go-header-file ./hack/boilerplate.go.txt
# then we need to edit the gen files, add module name (e.g. github.com/FFFFFaraway/MPI-Operator/
# make sure `go mod tidy` can run successfully
