FROM nvidia/cuda:11.3.1-devel-ubuntu20.04

ENV CUDNN_VERSION=8.2.0.53-1+cuda11.3
ENV NCCL_VERSION=2.9.9-1+cuda11.3

ENV TZ=Asia/Dubai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Set default shell to /bin/bash
SHELL ["/bin/bash", "-cu"]

RUN apt-get update --allow-insecure-repositories && apt-get install -y --allow-unauthenticated --allow-downgrades --allow-change-held-packages --no-install-recommends \
        build-essential \
        cmake \
        g++-9 \
        git \
        curl \
        vim \
        wget \
        ca-certificates \
        libcudnn8=${CUDNN_VERSION} \
        libnccl2=${NCCL_VERSION} \
        libnccl-dev=${NCCL_VERSION} \
        libjpeg-dev \
        libpng-dev \
        python-is-python3 \
        python3-pip \
        python3-dev \
        python3-distutils \
        librdmacm1 \
        libibverbs1 \
        ibverbs-providers

RUN pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

RUN pip install future typing packaging numpy https://download.pytorch.org/whl/cu113/torch-1.11.0%2Bcu113-cp38-cp38-linux_x86_64.whl

RUN HOROVOD_GPU_OPERATIONS=NCCL \
    HOROVOD_WITH_PYTORCH=1 \
    HOROVOD_WITHOUT_MPI=1 \
         pip install --no-cache-dir horovod

# Install OpenSSH for MPI to communicate between containers
RUN apt-get install -y --no-install-recommends openssh-client openssh-server && \
    mkdir -p /var/run/sshd

# Allow OpenSSH to talk to containers without asking for confirmation
RUN cat /etc/ssh/ssh_config | grep -v StrictHostKeyChecking > /etc/ssh/ssh_config.new && \
    echo "    StrictHostKeyChecking no" >> /etc/ssh/ssh_config.new && \
    mv /etc/ssh/ssh_config.new /etc/ssh/ssh_config
