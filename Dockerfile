FROM python:3.11.6-alpine

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip install -r requirements.txt

COPY main.py main.py

ENV TZ=Asia/Shanghai

CMD ["python", "-u", "main.py"]
