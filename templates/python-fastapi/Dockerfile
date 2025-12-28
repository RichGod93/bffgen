FROM python:3.11-slim

WORKDIR /app

# Copy requirements and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY src/ ./src/

# Expose port
EXPOSE {{PORT}}

# Run the application
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "{{PORT}}"]
