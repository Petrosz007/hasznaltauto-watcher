build-docker:
  docker build -t hasznaltauto-watcher:SNAPSHOT .

run-docker:
  docker run --rm -it -v ./:/data/ hasznaltauto-watcher:SNAPSHOT

release-docker VERSION:
  docker build -t andipeter/hasznaltauto-watcher:{{ VERSION }} .
  docker push andipeter/hasznaltauto-watcher:{{ VERSION }}
