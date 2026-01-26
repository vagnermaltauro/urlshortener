# 🚀 Como Usar o URL Shortener

## ✅ Sistema Rodando

Acesse: **http://localhost:8080/**

---

## 📋 Testando a Interface Web

1. **Abra no navegador:** http://localhost:8080/

2. **Cole uma URL** no campo de texto:
   ```
   https://github.com/anthropics/claude-code
   ```

3. **Clique em "Encurtar URL"**

4. **Resultado esperado:**
   - ✅ Mensagem de sucesso verde
   - 🔗 URL encurtada aparece abaixo (ex: http://localhost:8080/JWrBj6I9A0)
   - 📋 Botão "Copiar Link" para copiar a URL
   - 📅 Data de expiração (5 anos)

---

## 🔧 Se não aparecer o resultado:

### 1. Limpe o cache do navegador:
```
Chrome/Edge: Ctrl+Shift+Del
Firefox: Ctrl+Shift+Del
Safari: Cmd+Option+E
```

### 2. Ou force reload:
```
Chrome/Edge/Firefox: Ctrl+F5
Safari: Cmd+Shift+R
```

### 3. Abra o Console do navegador (F12) e veja se há erros

---

## 🧪 Testando via Terminal (cURL)

### Criar URL:
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/anthropics/claude-code"}'
```

**Resposta:**
```json
{
  "short_url": "http://localhost:8080/JWrBj6I9A0",
  "short_code": "JWrBj6I9A0",
  "expires_at": "2030-12-30T13:07:07Z"
}
```

### Testar Redirect:
```bash
curl -I http://localhost:8080/JWrBj6I9A0
```

**Resposta:**
```
HTTP/1.1 301 Moved Permanently
Location: https://github.com/anthropics/claude-code
```

---

## 📊 Verificar Métricas:
```bash
curl http://localhost:8080/metrics
```

---

## 🐛 Debug

### Ver logs do container:
```bash
docker logs url-shortener -f
```

### Ver status dos containers:
```bash
docker ps
```

### Reiniciar sistema:
```bash
docker-compose -f docker-compose.minimal.yml restart app
```

---

## ✨ Funcionalidades da Interface

- ✅ **Validação automática** de URLs
- ✅ **Feedback visual** (spinner durante processamento)
- ✅ **Mensagens de erro** claras
- ✅ **Copiar link** com um clique
- ✅ **Design responsivo** (funciona no celular)
- ✅ **Acessibilidade** (ARIA labels, foco automático)

---

## 🎯 Exemplo de Uso Completo

1. Acesse http://localhost:8080/
2. Cole: `https://www.youtube.com/watch?v=dQw4w9WgXcQ`
3. Clique "Encurtar URL"
4. Veja o resultado: `http://localhost:8080/JWr12AB34C`
5. Clique em "📋 Copiar Link"
6. Compartilhe a URL curta!
7. Quando alguém acessar, será redirecionado para o YouTube

---

## 💡 Dica

Se estiver com problema de cache do navegador, teste em **modo anônimo/privado**:
- Chrome: Ctrl+Shift+N
- Firefox: Ctrl+Shift+P
- Safari: Cmd+Shift+N
