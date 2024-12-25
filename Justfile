set shell := [ "bash", "-c" ]

NEW_VERSION := shell('echo $(date -I)-$(git rev-parse --verify HEAD)')
NEW_IMAGE_TAG := 'andipeter/hasznaltauto-watcher:' + NEW_VERSION

build-docker:
  docker build -t hasznaltauto-watcher:SNAPSHOT .

run-docker:
  docker run --rm -it -v ./:/data/ hasznaltauto-watcher:SNAPSHOT

release-docker:
  @echo 'New image tag: {{NEW_IMAGE_TAG}}'
  docker build -t {{NEW_IMAGE_TAG}} .
  docker push {{NEW_IMAGE_TAG}}
