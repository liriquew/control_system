FROM python:3.12

WORKDIR /app

COPY . ./

RUN pip install --no-cache-dir -r requirements.txt

ENV PYTHONUNBUFFERED=1

ENTRYPOINT ["python", "main.py"]
