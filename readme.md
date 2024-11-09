# Desafio: Rate Limiter em Go

## Objetivo
Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

## Descrição
O objetivo deste desafio é criar um rate limiter em Go que possa ser utilizado para controlar o tráfego de requisições para um serviço web. O rate limiter deve ser capaz de limitar o número de requisições com base em dois critérios:

1. **Endereço IP**: O rate limiter deve restringir o número de requisições recebidas de um único endereço IP dentro de um intervalo de tempo definido.

2. **Token de Acesso**: O rate limiter também pode limitar as requisições baseadas em um token de acesso único, permitindo diferentes limites de tempo de expiração para diferentes tokens. O Token deve ser informado no header no seguinte formato:
   - API_KEY: <TOKEN> 
   - As configurações de limite do token de acesso devem se sobrepor às do IP. Por exemplo, se o limite por IP for de 10 requisições por segundo e o de um token for de 100 requisições por segundo, o rate limiter deve utilizar as configurações do token.

## Requisitos

1. **Middleware**:
- O rate limiter deve ser implementado como um middleware que pode ser injetado ao servidor web.

2. **Limitação de Requisições**:
- O rate limiter deve permitir a configuração do número máximo de requisições permitidas por segundo.
- Deve ser possível configurar o tempo de bloqueio do IP ou do Token caso a quantidade de requisições seja excedida.

3. **Configurações via Variáveis de Ambiente**:
- As configurações de limite devem ser feitas através de variáveis de ambiente ou em um arquivo `.env` na pasta raiz.

4. **Respostas Adequadas**:
- O sistema deve responder adequadamente quando o limite for excedido:
    - Código HTTP: `429`
    - Mensagem: `you have reached the maximum number of requests or actions allowed within a certain time frame`

5. **Armazenamento em Redis**:
- Todas as informações de "limiter" devem ser armazenadas e consultadas de um banco de dados Redis.
- O Redis pode ser inicializado via Docker Compose.

6. **Estratégia de Persistência**:
- Crie uma “strategy” que permita trocar facilmente o Redis por outro mecanismo de persistência.

7. **Separação de Lógica**:
- A lógica do rate limiter deve estar separada do middleware.

## Exemplos

- **Limitação por IP**: Suponha que o rate limiter esteja configurado para permitir no máximo 5 requisições por segundo por IP. Se o IP `192.168.1.1` enviar 6 requisições em um segundo, a sexta requisição deve ser bloqueada.

- **Limitação por Token**: Se um token `abc123` tiver um limite configurado de 10 requisições por segundo e enviar 11 requisições nesse intervalo, a décima primeira deve ser bloqueada.

Nos dois casos, as próximas requisições poderão ser realizadas somente quando o tempo total de expiração ocorrer. Exemplo: Se o tempo de expiração for de 5 minutos, determinado IP ou token poderá realizar novas requisições somente após os 5 minutos.

## Dicas

- Teste seu rate limiter sob diferentes condições de carga para garantir que ele funcione conforme esperado em situações de alto tráfego.

## Entrega

- O código-fonte completo da implementação.
- Documentação explicando como o rate limiter funciona e como ele pode ser configurado.
- Testes automatizados demonstrando a eficácia e a robustez do rate limiter.
- Utilize `docker/docker-compose` para facilitar a execução da aplicação.
- O servidor web deve responder na porta `8080`.

## Explicação do projeto

1. **Recebendo a requisição**: Quando uma requisição chega ao sistema, o rate limiter é acionado.

2. **Configuração do middleware**: O middleware processa a requisição e carrega as configurações de rate limit (número máximo de requisições por segundo) e block duration (tempo de bloqueio em minutos) a partir das variáveis de ambiente. Caso um token de acesso seja fornecido, o rate e o tempo de bloqueio podem ser alterados conforme as informações contidas nesse token.

   Exemplo de token: `token_rate_1000_blockduration_1`

3. **Processamento no caso de uso**: O middleware encaminha a requisição para o serviço de rate limiting, que executa as seguintes ações:

    - Se o usuário ainda não foi registrado, o sistema cria uma nova entrada para ele, contando o primeiro acesso.
    - Caso o usuário já tenha uma entrada, o sistema verifica se o tempo de expiração passou. Se o limite de tempo (1 segundo) foi alcançado, o contador de acessos é reiniciado.
    - Se o usuário não ultrapassou o limite de requisições, o contador de acessos é incrementado.
    - Se o limite de requisições for excedido, o sistema bloqueia o usuário pelo tempo configurado.

4. **Troca de persistência**: Para tornar o sistema flexível, foi implementada uma estratégia que permite substituir o Redis por outro mecanismo de persistência. Isso é possível porque a arquitetura segue os princípios da Clean Architecture, permitindo que o repositório de dados seja facilmente alterado, sem impacto nas demais partes do sistema.

## Configuração do projeto

1. **Clone o Repositório:**

   ```bash
   git clone https://github.com/vinicius-maker/rate-limiter.git
   cd rate-limiter

2. **Configurar variáveis de ambiente (opcional):**
    - no caminho rate-limiter/cmd
    - já está com configurações padrões
    - obs: se alterar os valores padrões do arquivo .env, os testes apresentarão um comportamento diferente, devido às configurações dos valores padrões

3. **Configurar docker:**
    - no diretório raiz: rate-limiter/

    ```bash
        docker-compose build
        docker-compose up -d app-prod

4. **Acessar o localhost:**
    - Após a execução do Docker, o servidor estará disponível em http://localhost:8080/. Acesse essa URL no seu navegador para interagir com a aplicação.

5. **Testes:**

   #### Testar Comportamento no Navegador

   - Para testar a limitação de requisições diretamente no navegador, basta atualizar a página 6 vezes consecutivas, em menos de 1 segundo. A sexta requisição será bloqueada, conforme a configuração de rate limit.

   #### Testes de Carga com ApacheBench

   Para realizar testes de carga e verificar o comportamento do rate limiter, você pode utilizar o ApacheBench. Abaixo estão alguns exemplos:

   - **teste com 10 requisições, com 1 conexão por IP:**

    ```bash
    ab -n 10 -c 1 -k http://localhost:8080/
    ```
   
   - obs: caso realizado o teste via navegador, aguardar 1 minuto para executar o teste via ApacheBench (block_duration por padrão está configurado 1 minuto)

   - **teste com 1001 requisições, com 1 conexão via TOKEN**
    
    ```bash
    ab -n 1001 -c 1 -k -H "API_KEY: token_rate_1000_blockduration_1"  http://localhost:8080/
    ```
     - obs: no token é específicado o rate limit e o tempo de bloqueio
     - obs2: se for utilizar a ferramenta, se atentar aos campos "Complete requests" (sucesso) e "Non-2xx responses" (erro) gerados nos logs para apuração dos resultados

   #### Testar via testes unitários
    - será necessário entrar do docker
    ```bash
    docker exec -it app-prod bash
    go test ./...
    ```
    - obs: se alterar os valores padrões do arquivo .env, os testes apresentarão um comportamento diferente, devido às configurações dos valores padrões