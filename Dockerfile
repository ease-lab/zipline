FROM python:slim as pygx

RUN apt update && \
    apt install g++ make cmake -y && \
    pip install pycapnp grpcio && \
    apt remove g++ make cmake -y && \
    apt autoremove -y
RUN pip install environs grpcio-tools
WORKDIR /app
COPY sdk/python/ ./
COPY user-functions/gx/gx.py ./
COPY ./proto/*/*.py ./
COPY ./proto/crossXDT/crossXDT.py.capnp ./

ENTRYPOINT ["python3", "-u", "gx.py"]

FROM python:slim as pyfx

RUN apt update && \
    apt install g++ make cmake -y && \
    pip install pycapnp grpcio && \
    apt remove g++ make cmake -y && \
    apt autoremove -y
RUN pip install environs grpcio-tools
WORKDIR /app
COPY sdk/python/ ./
COPY user-functions/fx/fx.py ./
COPY ./proto/*/*.py ./
COPY ./proto/crossXDT/crossXDT.py.capnp ./

ENTRYPOINT ["python3", "-u", "fx.py"]