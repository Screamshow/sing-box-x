# Sing-Box X 1.0.3: локальный кандидат XHTTP, 2026-10-08

## Результат и границы

Подготовлен локальный кандидат X 1.0.3 на upstream 1.14.2: executable
`1.14.2-x-1.0.3`, APK `1.0.3-r1`, IPK `1.0.3-1`. Добавлен клиент XHTTP
HTTP/1.1 и HTTP/2. HTTP/3 и QUIC не включены. Релиз не опубликован, зеркало
не обновлялось. APK не подписаны. Это проверенный тестовый кандидат, а не
заявление о совместимости со всеми конфигурациями Xray или о длительной нагрузке.

Оба рабочих репозитория остались на существующих main. Исходные незакоммиченные
изменения 1.0.2 сохранены; его исходники, сборки, dist и адаптеры оставлены
в `work/source-1.0.2`, `work/baseline-1.0.2` и `work/dist-before-xhttp`.
Восемь исправлений устойчивости и освобождение состояния uTLS применяются перед
XHTTP. Файлы `common/tls/utls_client.go` и `common/tls/reality_client.go`
в подготовленных исходниках побайтово совпадают с кандидатом 1.0.2.
Отпечатки TLS браузеров, REALITY client version и sing-vmess не менялись.
Предыдущие RAM-замеры 64 REALITY/Vision/Firefox соединений не повторялись.

## Перенос и происхождение

Оригинальные патчи FiyeroT из podkop-engine v1.14.2-r12 (46e2370) сохранены
побайтово с заголовками авторства и лицензией GPL-3.0-or-later в `patches/xhttp`.
SHA-256 и порядок приведены в `patches/xhttp/manifest.json`.
Перенесены 0046–0054, 0057–0064, 0069, 0071–0073; из 0068 взято объявление
возможностей. Зависимость 0031 применяется только к `common/readwait/*`.
Уже закреплённый cpuid v2.3.0 стал прямой зависимостью; посторонние forks go.mod
не перенесены. TLS/REALITY-патчи, меняющие fingerprint или client version,
для этого набора не понадобились.

Адаптер `scripts/prepare-xhttp.py` проверяет входные хеши, применяет последующие
исправления, удаляет реализацию HTTP/3 и QUIC TLS helpers, сохраняет явный отказ
от HTTP/3. Исключительный ALPN h3 отвергается и на основной, и на download-ветви,
включая REALITY. Два HTTP/2 регрессионных теста из смешанного review2 включены
без with_quic. XHTTP-заголовки HTTP следуют донорской реализации; это не меняет
TLS ClientHello пользователя.

Обнаружена гонка первой HTTP/2 upload-записи с получением SETTINGS:
MAX_HEADER_LIST_SIZE мог ещё не быть известен, и сервер закрывал соединение
через GOAWAY вместо ошибки отдельной сессии. Добавлен pool с SETTINGS/PING
барьером перед выдачей нового ClientConn. Регрессия повторяется на 20 свежих
соединениях. В донорском тесте сохранения соединения XMUX ограничен одним
соединением, чтобы проверка не зависела от ленивого создания трёх сокетов.

Подготовленные исходники содержат `.forkop-xhttp.json` и
`.forkop-xhttp-effective.json`: входные патчи, SHA адаптера, helper, тестов и
эффективных файлов транспорта. Source archive дополнительно содержит
`.forkop-provenance` с оригинальными патчами, адаптерами, тестами, конфигурацией
сборки, packaging и лицензией. Подготовка повторно проверена на идемпотентность.

## Forkop и пользовательский фильтр

Runtime и UI определяют XHTTP по точному токену `transport.xhttp` в строке
`Features:` вывода version. Новый X объявляет также `transport.xhttp.http1`
и `transport.xhttp.http2`. Имя/номер X сами по себе не означают XHTTP.
Для старых Extended и Extended compressed сохранена прежняя совместимость;
обычный sing-box, Tiny и X без feature остаются без XHTTP. UI передаёт
`sing_box_xhttp`, frontend использует его и совместимый fallback для старого
backend. Кеш версии без поля features перепроверяется. Генератор получает
возможность независимо от признака Extended.

Пользовательский фильтр белого флага сохранён. Актуальная подписка загружена
штатным механизмом Forkop с INCY и автоматически полученным HWID и нормализована
штатным parser: **62 записи outbound**, включая одну группу URLTest.
Самих серверов **61: 58 TCP REALITY, 1 gRPC и 2 XHTTP**. После действующего
фильтра осталось **58 TCP**, 0 gRPC, 0 XHTTP. Предыдущий подсчёт 59 TCP ошибочно
включал группу URLTest как TCP-сервер. Включение возможности XHTTP результат
фильтра не меняет.
Для изолированного протокольного теста два исключённых XHTTP узла использованы
явно, вне пользовательской маршрутизации. Ссылка, HWID, имена серверов и
учётные данные не включены в отчёт, provenance или артефакты кандидата.

## Тесты и сборка

Unit и race пройдены для route/rules, Clash API, sniff/ja3/interrupt/readwait,
TLS, DNS/DoH, log, gRPC-lite, HTTP, WebSocket/HTTPUpgrade и XHTTP.
Полный XHTTP набор: unit 50,990 с, race 55,557 с. Покрываются режимы, download
leg, ошибочные параметры/статусы, отказ dial, бесконечное тело ошибки,
разблокировка upload, размеры frame/header, deadlines, XMUX rotation/churn,
reset после ошибок и во время dial, отмена контекста и закрытие retired clients.
Добавлены проверки auto REALITY/download и отказа h3 на обеих ветвях.

Сборка CGO=0, Go 1.26.8, UPX 5.2.1 для ARM64 и x86_64. Теги:
`with_utls,with_clash_api,without_native_api,badlinkname,tfogo_checklinkname0`.
Графы зависимостей обоих архитектур проверены: XHTTP присутствует, отсутствуют
quic-go, sing-quic, Hysteria/Hysteria2/TUIC, gVisor, отдельный native API и
google.golang.org/grpc. Сохранены Clash API и gRPC-lite. Сборщик отвергает
with_quic. UPX integrity, ELF architecture, executable mode и package layout
проверены. ARM64 дополнительно выполнена на пользовательском GL-MT6000
192.168.90.1: start/restart, реальная подписка, DNS, списки и Clash API прошли.
Подробности: `docs/sing-box-x-xhttp-router90-2026-10-08.md` в основном Forkop repo.

Forkop: regression runtime/capability/ruleset-cache пройдены; parser на VM —
308 assertions; frontend — 15 тестов в двух файлах, TypeScript и форматирование.

## Прирост относительно 1.0.2

Значения — байты файлов и извлечённого payload, не размер архива и не размер
UPX-декомпрессии в RAM. Исходные пакеты сравниваются с сохранённым dist 1.0.2.

| Показатель | x86_64: 1.0.2 → 1.0.3 | ARM64: 1.0.2 → 1.0.3 | Прирост |
|---|---:|---:|---:|
| UPX binary | 9 868 792 → 9 934 048 | 8 541 740 → 8 599 568 | +65 256 / +57 828 |
| Plain executable | 31 158 434 → 31 338 658 | 28 836 002 → 29 032 610 | +180 224 / +196 608 |
| IPK extracted payload | 9 870 931 → 9 936 187 | 8 543 891 → 8 601 719 | +65 256 / +57 828 |
| APK installed payload | 9 871 066 → 9 936 322 | 8 544 026 → 8 601 854 | +65 256 / +57 828 |

UPX binary вырос примерно на 0,66% / 0,68%. Зависимости пакетов остаются
ca-bundle и kmod-tun. Эта таблица не заменяет flash-планирование установки с
rollback-пакетами и временным workspace.

## Интеграция и реальные узлы

Использованы существующие VMware OpenWrt 25 и 24. До изменений сохранены
пакеты, состояния/включение сервисов, конфигурация и контрольные суммы.
Локальные пакеты не устанавливались: только временная bind-подмена бинарника
и изменённых backend файлов. Перед отдельными процессами sing-box Forkop
останавливался, чтобы watchdog не вмешивался.

Матрица на OpenWrt 25 с контролируемым Xray:

| Транспорт | auto | stream-one | stream-up | packet-up | downloadSettings |
|---|---|---|---|---|---|
| HTTP/1.1 plain | passed | passed | passed | passed | — |
| HTTP/1.1 TLS | passed | passed | passed | passed | — |
| HTTP/2 TLS | passed | passed | passed | passed | auto/stream-up/packet-up passed |
| HTTP/2 REALITY, Firefox, Xray 25.8.3 | passed | passed | passed | passed | auto/stream-up/packet-up passed |

22 успешных сочетания. downloadSettings использует независимый клиентский pool
с общим серверным XHTTP session endpoint. stream-one с отдельной download-ветвью
не является допустимым сочетанием. Unit-тесты дополнительно проверяют варианты
и ошибки downloadSettings. Первые попытки с HTTP fixture, недоступной uhttpd из-за
прав, и отдельными Xray inbound без общих sessions были исправлены и не засчитаны
как транспортные результаты.

Два узла реальной подписки используют packet-up TLS, Firefox/Chrome, без
отдельного downloadSettings; оба прошли HTTPS запросы: gstatic 204 и example
200. HTTP/3 для них не понадобился. Это функциональная проверка текущих двух
узлов, не throughput/soak-тест.

Дополнительно выполнена свежая массовая проверка тем же Clash API `/delay`,
который использует пинг дашборда, на кандидате 1.0.3: три последовательных
прохода **61/61, 61/61, 61/61**, всего **183 успешных URLTest без ошибок**.
Каждый проход включал 58 TCP REALITY, один gRPC REALITY и два XHTTP TLS.
URLTest открывает выбранный outbound и делает HTTPS HEAD к gstatic через него,
проверяя TLS/REALITY и получение HTTP-ответа; это не ICMP и не тест пропускной
способности. В сам URLTest не заложена проверка конкретного HTTP status.
Пользовательский фильтр в штатной конфигурации не менялся: исключённые узлы
проверены в отдельном процессе после остановки Forkop. На VM восстановлены
исходный X 1.0.1, UCI/generated config и сервис; пакеты не менялись, временные
файлы удалены. Результаты: `work/xhttp-evidence/subscription-latency-3rounds.json`
и `subscription-latency-restoration.log`. Это подтверждает доступность всех
серверов текущей подписки с X 1.0.3 на момент проверки, включая её gRPC;
не является общей гарантией всех вариантов gRPC или новых REALITY серверов.

REALITY с новыми контролируемыми Xray 26.9.8/26.9.30 не прошёл авторизацию:
после fallback возникала ошибка доверия сертификату. Контроль со штатным
X 1.0.1 на TCP REALITY и Xray 26.9.8 воспроизвёл тот же сбой. Он не доказывает
новую регрессию XHTTP, но ограничивает совместимость кандидата: последняя
постквантовая REALITY не заявляется. Fingerprint/client version не менялись
для обхода сбоя. Для новых серверных REALITY параметров нужен отдельный перенос
и тестирование.

На обеих VM кандидат прошёл start/restart Forkop, DNS и Clash API. На OpenWrt 25
полная генерация с действующим фильтром дала 58 VLESS outbounds и один ruleset;
UI показал XHTTP=1 и Extended=0. На OpenWrt 24 проверялся штатный минимальный
набор конфигурации. Отдельно на обеих VM пройдена валидация SRS с 10 003 доменами:
large/nested, malformed/truncated/unsupported/changed, cache publication,
unchanged refresh, last-known-good и RAM-only cache с последующей promotion.
На OpenWrt 25 JSON и SRS списки реально направили HTTPS через XHTTP; Clash API
подтвердил rule_set и целевой host. Настройка Disable QUIC не отключалась.

После тестов обе VM возвращены к исходному бинарнику/конфигурации. Проверены
хеши бинарника, Forkop/sing-box/DHCP UCI, generated config, runtime/UI/generator,
init script, равенство списка пакетов, работа/включение Forkop, DNS и Clash API.
Bind-подмены сняты. Временные конфиги подписки, серверные ключи и тестовые
процессы удалены с VM после сохранения обезличенных результатов.

## Локальные артефакты и воспроизведение

`dist/` содержит семь артефактов: два UPX binary tar.gz, два APK, два IPK,
source tar.gz, плюс manifest.json и SHA256SUMS. Для всех проверены длина,
SHA-256, package identity и соответствие binary. URL в manifest — шаблоны
будущего размещения, не существующая публикация и не инструкция установки.

Сборка: `APK_TOOL=<OpenWrt SDK apk> bash scripts/build.sh` с указанным Go.
Она требует чистый work/source, применяет hardening затем XHTTP, тестирует,
проверяет графы, собирает/упаковывает и сохраняет provenance. Уже готовый
кандидат находится в dist; текущие baseline/source каталоги удалять не нужно.
VM harness: `tests/xhttp-matrix.sh`/`.uc` и `tests/subscription-probe.sh`;
запускать после сохранения VM состояния и остановки Forkop.

Доказательства сборки: `work/build-xhttp.log`, графы `work/dependencies*.txt`;
обезличенные VM результаты: `work/xhttp-evidence/`, ARM64 router:
`work/xhttp-evidence/router90/`. Ограничения: новые REALITY серверы ограничены,
HTTP/3 отсутствует, пакетная установка/откат кандидата ещё не проверены, долгосрочная
нагрузка и расширенные топологии downloadSettings требуют отдельных тестов.
