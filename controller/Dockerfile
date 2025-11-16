# Python Controller Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Make CLI executable
RUN chmod +x main.py

# Default command
ENTRYPOINT ["python", "main.py"]
CMD ["--help"]
