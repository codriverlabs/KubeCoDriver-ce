mkdir -p ./tmp
rm -rf ./tmp/*
rm -rf ./dist/*
export VERSION=1.0.43
make manifests
make helm-chart
make helm-package
cp ./dist/helm/*.tgz ./tmp/
cd ./tmp
tar -xf ./*.tgz
AWS_ACCOUNT_ID=864899852480
AWS_REGION=ap-south-1
helm template --debug \
--set-string global.registry.repository=864899852480.dkr.ecr.ap-south-1.amazonaws.com/codriverlabs/kubecodriver \
--set-string controller.image.tag=$VERSION \
--set-string collector.image.tag=$VERSION kubecodriver-operator-$VERSION ./kubecodriver-operator > template.yaml
cd ..
