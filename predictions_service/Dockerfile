FROM python:3.12

WORKDIR /app

COPY . .

RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 4041

ENV PYTHONUNBUFFERED=1

CMD ["python", "main.py"]
