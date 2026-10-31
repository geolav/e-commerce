#!/bin/bash

# Создаём zip архив, исключая ненужные файлы и папки
zip -r projectMyTests.zip . \
  -x "*.git/*" \
  -x "*node_modules/*" \
  -x "*/mocks/*" \
  -x "*mock*" \
  -x "*bin*" \
  -x "*vendor*" \
  -x "*.exe" \
  -x "*.dll" \
  -x "*.so" \
  -x "*.o" \
  -x "*.a" \
  -x "*.out" \
  -x "*.test" \
  -x "*generated*" \
  -x "*/frontend/*" \
  -x "*/frontend-*/*" \
  -x "*.zip" \
  -x "*.tar.gz" \
  -x "*.log" \
  -x "*/tmp/*" \
  -x "*/temp/*" \
  -x "*__pycache__/*" \
  -x "*.pyc" \
  -x ".DS_Store" \
  -x "Thumbs.db" \
  -x "*/test-output.log"

echo "Архив project.zip создан. Размер: $(du -h project.zip | cut -f1)"
