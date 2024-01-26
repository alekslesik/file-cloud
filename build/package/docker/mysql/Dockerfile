# syntax=docker/dockerfile:1

# set base img
FROM mysql

# add metadata to img
LABEL description="MysSQL database for file-cloud application"
LABEL maintainer="github.com/alekslesik"
LABEL app="alekslesik/file-cloud"

# set work dir for above instructions
WORKDIR /settings

ENV MYSQL_ROOT_PASSWORD=Todor1990///

# # execute a command and create slice of img
# RUN apt-get update && apt-get upgrade \
#     package-one \
#     package-two \
#     package-three && rm -rf /var/lib/apt/lists/*


# copy files to container
# COPY /root/go/src/github.com/alekslesik/file-cloud/.envrc .

# COPY ../../migrations .

# # describe the command with arguments which must be executed when the container will be started
# # the arguments can be redefine when cantainer run
# # there is only one instruction in the DockerFile
# CMD ["node", "src/index.js"]

# point to need open a port
EXPOSE 3306

# # create storage to save data
# VOLUME /my_volume