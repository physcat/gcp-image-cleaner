# Clean up old docker image on gcp


When setting up a cloud build for every branch, there are a lot of docker images left over.

This command will remove all docker images older than 90 days and will ignore images with a sem version tag. e.g. `v0.1.0-test`
