FROM golang:1.15

ARG binary_filename=bot
ARG project_path=/src/${binary_filename}

COPY . ${project_path}
WORKDIR ${project_path}

RUN go get -d -v ./...
RUN go install -v ./...

ARG binary_dir=${project_path}/src/main
ENV binary_filepath ${binary_dir}/${binary_filename}

RUN go build -o ${binary_filepath} ${binary_dir}

CMD $binary_filepath "$config_path"
