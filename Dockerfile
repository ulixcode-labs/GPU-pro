FROM nvidia/cuda:13.0.1-cudnn-devel-ubuntu24.04
COPY gpu-pro /usr/local/bin/
EXPOSE 1312
CMD ["/usr/local/bin/gpu-pro"]
