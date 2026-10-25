<?php

declare(strict_types=1);

/**
 * Code generator for the CryptoDataDownload Go SDK.
 *
 * Reads the live OpenAPI (Swagger 2.0) schema and emits:
 *   - zz_gen_<service>.go        one service file per schema tag (21 services)
 *   - zz_contract_test.go        per-operation coverage test (all 116 ops)
 *   - testdata/operations.json   manifest (used by docs + future checks)
 *   - docs/api-coverage.md       human coverage table with doc links
 *
 * Usage:  php tools/generate.php [path/to/openapi.json]
 */

$pkg = dirname(__DIR__);
$input = $argv[1] ?? null;
if ($input === null) {
    foreach ([$pkg . '/testdata/openapi.json', $pkg . '/../../../cdd-snapshot/cdd_openapi.json', $pkg . '/cdd_openapi.json'] as $cand) {
        if (is_file($cand)) { $input = $cand; break; }
    }
}
if ($input === null) {
    fwrite(STDERR, "schema not found; pass path to openapi json\n");
    exit(1);
}
$schemaSource = (string) file_get_contents($input);
$schema = json_decode($schemaSource, true, 512, JSON_THROW_ON_ERROR);
// Preserve JSON objects in response fixtures, including empty objects.
$schemaObjects = json_decode($schemaSource, false, 512, JSON_THROW_ON_ERROR);
$baseURL = 'https://' . $schema['host'] . $schema['basePath'];
$uiURL = 'https://' . $schema['host'] . '/';
$schemaURL = $uiURL . '?format=openapi';
$catalogURL = $uiURL . 'static/catalog/CDD_API_Endpoints.csv';
$catalog = [];
$catalogFile = $argv[2] ?? $pkg . '/testdata/endpoints.csv';
if (is_file($catalogFile)) {
    $handle = fopen($catalogFile, 'r');
    $columns = fgetcsv($handle, 0, ',', '"', '');
    while (($row = fgetcsv($handle, 0, ',', '"', '')) !== false) {
        if (count($row) === count($columns)) {
            $entry = array_combine($columns, $row);
            $catalog[$entry['Endpoint']] = $entry;
        }
    }
    fclose($handle);
}

// tag slug -> [ServiceType, client field]
$services = [
    'aggregates'                    => ['AggregatesService', 'Aggregates'],
    'binance-ohlc'                  => ['BinanceOHLCService', 'BinanceOHLC'],
    'binance-summary'               => ['BinanceSummaryService', 'BinanceSummary'],
    'deribit-ohlc'                  => ['DeribitOHLCService', 'DeribitOHLC'],
    'deribit-summary'               => ['DeribitSummaryService', 'DeribitSummary'],
    'catalog-coverage'              => ['CatalogService', 'Catalog'],
    'market-breadth-indicators'     => ['BreadthService', 'Breadth'],
    'on-chain-data'                 => ['OnChainService', 'OnChain'],
    'options-analytics'             => ['OptionsAnalyticsService', 'OptionsAnalytics'],
    'options-pricing-tools'         => ['OptionsPricingService', 'OptionsPricing'],
    'perp-funding-rates'            => ['FundingService', 'Funding'],
    'pre-computed-ml-features-cross-sectional'      => ['MLCrossSectionalService', 'MLCrossSectional'],
    'pre-computed-ml-features-futures-positioning'  => ['MLFuturesService', 'MLFutures'],
    'pre-computed-ml-features-macro-signals'        => ['MLMacroService', 'MLMacro'],
    'pre-computed-ml-features-market-basics'        => ['MLMarketService', 'MLMarket'],
    'pre-computed-ml-features-microstructure'       => ['MLMicrostructureService', 'MLMicrostructure'],
    'pre-computed-ml-features-options-implied'      => ['MLOptionsService', 'MLOptions'],
    'pre-computed-ml-features-regime-anomaly'       => ['MLRegimeService', 'MLRegime'],
    'risk-volatility-metrics'       => ['RiskService', 'Risk'],
    'technical-indicators'          => ['TechnicalsService', 'Technicals'],
    'us-govt-data'                  => ['GovtService', 'Govt'],
];

// wire-name -> Go field name overrides for readability
$fieldNameMap = [
    'enddate' => 'EndDate', 'startdate' => 'StartDate',
    'end_date' => 'EndDate', 'start_date' => 'StartDate',
    'group_by' => 'GroupBy', 'contract_name' => 'ContractName',
    'group_name' => 'GroupName', 'market_type' => 'MarketType',
    'days_to_maturity' => 'DaysToMaturity', 'min_usd' => 'MinUSD',
    'min_premium_usd' => 'MinPremiumUSD', 'min_usd_value' => 'MinUSDValue',
    'putcall' => 'PutCall', 'buysell' => 'BuySell', 'rfr' => 'RFR', 'iv' => 'IV',
];

function tagSlug(string $tag): string
{
    return strtolower(trim(preg_replace('/[^a-zA-Z0-9]+/', '-', $tag), '-'));
}

function camel(string $s): string
{
    $parts = preg_split('/[^a-zA-Z0-9]+/', $s, -1, PREG_SPLIT_NO_EMPTY);
    $first = array_shift($parts);
    $out = lcfirst($first ?? '');
    foreach ($parts as $p) { $out .= ucfirst($p); }
    return $out;
}

function pascal(string $s): string
{
    return ucfirst(camel($s));
}

function fieldName(string $wire): string
{
    global $fieldNameMap;
    return $fieldNameMap[$wire] ?? pascal($wire);
}

function goScalarType(array $p): string
{
    return match ($p['type'] ?? 'string') {
        'integer' => 'int',
        'number'  => 'float64',
        default   => 'string',
    };
}

function isDateParam(array $p): bool
{
    return ($p['format'] ?? '') === 'date' || ($p['name'] ?? '') === 'date';
}

function ptr(string $path): string
{
    return str_replace(['~', '/'], ['~0', '~1'], $path);
}

function oneLine(string $s): string
{
    $s = preg_replace('/\s+/', ' ', trim($s));
    return str_replace('*/', '* /', $s);
}

function markdownCell(string $value): string
{
    return str_replace(['|', "\r", "\n"], ['&#124;', '', '<br>'], $value);
}

function parameterSummary(array $params): string
{
    $parts = [];
    foreach ($params as $p) {
        if ($p['name'] === 'auth_token') continue;
        $text = '`' . $p['name'] . '` (' . (!empty($p['required']) ? 'required' : 'optional') . ', `' . $p['type'] . '`';
        if (isset($p['format'])) $text .= ', ' . $p['format'];
        if (array_key_exists('default', $p)) $text .= ', default `' . json_encode($p['default'], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . '`';
        if (isset($p['enum'])) $text .= ', enum `' . implode(', ', $p['enum']) . '`';
        $parts[] = markdownCell($text . ')');
    }
    return $parts ? implode('<br>', $parts) : '—';
}

// ---- determine result shape per op from schema examples ----
function resultShape(mixed $ex): string
{
    if ($ex instanceof stdClass && property_exists($ex, 'result')) {
        if (is_array($ex->result) && $ex->result !== []
            && count(array_filter($ex->result, 'is_string')) === count($ex->result)) {
            return 'strings';
        }
        return is_array($ex->result) ? 'rows' : 'object';
    }
    if ($ex instanceof stdClass) {
        return 'object'; // envelope-less bare object
    }
    return 'rows'; // no example: legacy list endpoints
}

// ---- group operations by tag ----
$groups = [];
foreach ($schema['paths'] as $path => $item) {
    $op = $item['get'];
    $slug = tagSlug($op['tags'][0] ?? 'uncategorized');
    if (!isset($services[$slug])) {
        fwrite(STDERR, "unmapped tag '{$op['tags'][0]}' for {$path}\n");
        exit(1);
    }
    $groups[$slug][] = [$path, $op];
}

$paramsUsed = []; // params struct name -> service (collision check)
$manifest = [];
$mdRows = [];
$total = 0;

foreach ($groups as $slug => $ops) {
    [$type, $field] = $services[$slug];

    $segLists = array_map(fn($o) => array_values(array_filter(explode('/', trim($o[0], '/')), 'strlen')), $ops);
    $prefix = $segLists[0];
    foreach ($segLists as $sl) {
        $i = 0;
        while ($i < count($prefix) && $i < count($sl) && $prefix[$i] === $sl[$i]) { $i++; }
        $prefix = array_slice($prefix, 0, $i);
    }

    $used = [];
    $methods = [];
    foreach ($ops as [$path, $op]) {
        $segs = array_values(array_filter(explode('/', trim($path, '/')), 'strlen'));
        $suffix = array_slice($segs, count($prefix));
        if (empty($suffix)) { $suffix = [end($segs)]; }
        if (count($ops) === 1 && count($suffix) === 1 && count($segs) > 2) {
            $suffix = array_slice($segs, -2);
        }
        $name = pascal(implode('-', $suffix));
        if ($name === '' || ctype_digit($name[0])) { $name = 'N' . $name; }
        if (isset($used[$name])) {
            $i = 2;
            while (isset($used[$name . $i])) { $i++; }
            $name .= $i;
        }
        $used[$name] = true;

        $paramsName = $name . 'Params';
        if (isset($paramsUsed[$paramsName]) && $paramsUsed[$paramsName] !== $type) {
            $paramsName = $type . $paramsName;
        }
        $paramsUsed[$paramsName] = $type;

        $required = [];
        $optional = [];
        $hasReturn = false;
        foreach (($op['parameters'] ?? []) as $p) {
            if ($p['name'] === 'auth_token') continue;
            if ($p['name'] === 'return') { $hasReturn = true; continue; }
            $entry = [
                'wire'    => $p['name'],
                'field'   => fieldName($p['name']),
                'goType'  => isDateParam($p) ? 'time.Time' : goScalarType($p),
                'isDate'  => isDateParam($p),
                'enum'    => $p['enum'] ?? null,
                'default' => $p['default'] ?? null,
                'desc'    => oneLine($p['description'] ?? ''),
                'isFields'=> $p['name'] === 'fields',
                'paginationMax' => null,
            ];
            if ($entry['goType'] === 'int' && in_array($p['name'], ['limit', 'offset'], true)) {
                $entry['paginationMax'] = (int) ($p['maximum'] ?? 0);
                if ($entry['paginationMax'] === 0 && preg_match('/\bmax\s*[=:]?\s*(\d+)/i', $entry['desc'], $match)) {
                    $entry['paginationMax'] = (int) $match[1];
                }
            }
            if (!empty($p['required'])) { $required[] = $entry; }
            else { $optional[] = $entry; }
        }

        $example = $schemaObjects->paths->{$path}->get->responses->{'200'}->examples->{'application/json'} ?? null;
        $methods[] = [
            'name' => $name, 'params' => $paramsName, 'path' => $path, 'op' => $op,
            'required' => $required, 'optional' => $optional,
            'hasReturn' => $hasReturn, 'shape' => resultShape($example), 'example' => $example,
        ];
        $total++;
    }
    $groups[$slug] = $methods;
}

// ---- emit service files ----
foreach ($groups as $slug => $methods) {
    [$type, $field] = $services[$slug];
    $tagName = null;
    foreach ($schema['paths'] as $p => $item) {
        if (tagSlug($item['get']['tags'][0] ?? '') === $slug) { $tagName = $item['get']['tags'][0]; break; }
    }

    $usesTime = $usesStrings = $hasRaw = false;
    foreach ($methods as $m) {
        foreach (array_merge($m['required'], $m['optional']) as $p) {
            if ($p['isDate']) { $usesTime = true; }
            if ($p['isFields']) { $usesStrings = true; }
        }
        if ($m['hasReturn']) { $hasRaw = true; }
    }

    $out = "// Code generated by tools/generate.php from the CryptoDataDownload live\n";
    $out .= "// OpenAPI schema ({$schemaURL}). DO NOT EDIT — regenerate instead.\n\n";
    $out .= "package cryptodatadownload\n\nimport (\n\t\"context\"\n\t\"net/url\"\n";
    if ($usesStrings) { $out .= "\t\"strings\"\n"; }
    if ($usesTime) { $out .= "\t\"time\"\n"; }
    $out .= ")\n\n";
    $out .= "// {$type} covers the \"{$tagName}\" operations of the CryptoDataDownload API.\n//\n// See {$uiURL}\ntype {$type} struct {\n\tclient *Client\n}\n\n";

    foreach ($methods as $m) {
        $path = $m['path'];
        $doc = $schemaURL . '#/paths/' . ptr($path) . '/get';
        $summary = oneLine($m['op']['summary'] ?? $m['name']);
        $opName = lcfirst($field) . '.' . lcfirst($m['name']);

        // ---- Params struct ----
        $out .= "// {$m['params']} are the parameters for {$type}.{$m['name']}.\n//\n// See {$doc}\ntype {$m['params']} struct {\n";
        foreach ($m['required'] as $p) {
            $def = $p['default'] !== null ? " Default \"{$p['default']}\"." : '';
            $out .= "\t// {$p['desc']}{$def}\n\t{$p['field']} {$p['goType']}\n";
        }
        foreach ($m['optional'] as $p) {
            $t = $p['isFields'] ? '[]string' : '*' . $p['goType'];
            $def = $p['default'] !== null ? " Default \"{$p['default']}\" (applied server-side when nil)." : '';
            $out .= "\t// {$p['desc']}{$def}\n\t{$p['field']} {$t}\n";
        }
        $out .= "}\n\n";

        // ---- query building ----
        $qbody = "\tq := url.Values{}\n";
        foreach ($m['required'] as $p) {
            $lv = lcfirst($p['field']);
            if ($p['isDate']) {
                $qbody .= "\tif err := requireDate(\"{$p['wire']}\", p.{$p['field']}); err != nil {\n\t\treturn nil, err\n\t}\n";
                $qbody .= "\tq.Set(\"{$p['wire']}\", dateString(p.{$p['field']}))\n";
            } elseif ($p['enum'] && $p['goType'] === 'string') {
                $list = implode(', ', array_map(fn($e) => '"' . $e . '"', $p['enum']));
                if ($p['default'] !== null) {
                    $qbody .= "\t{$lv} := p.{$p['field']}\n\tif {$lv} == \"\" {\n\t\t{$lv} = \"{$p['default']}\"\n\t}\n";
                    $qbody .= "\tif err := checkEnum(\"{$p['wire']}\", {$lv}, {$list}); err != nil {\n\t\treturn nil, err\n\t}\n";
                    $qbody .= "\tq.Set(\"{$p['wire']}\", {$lv})\n";
                } else {
                    $qbody .= "\tif err := checkEnum(\"{$p['wire']}\", p.{$p['field']}, {$list}); err != nil {\n\t\treturn nil, err\n\t}\n";
                    $qbody .= "\tq.Set(\"{$p['wire']}\", p.{$p['field']})\n";
                }
            } elseif ($p['goType'] === 'int' || $p['goType'] === 'float64') {
                $value = "p.{$p['field']}";
                if ($p['default'] !== null) {
                    $qbody .= "\t{$lv} := p.{$p['field']}\n\tif {$lv} == 0 {\n\t\t{$lv} = {$p['default']}\n\t}\n";
                    $value = $lv;
                }
                if ($p['goType'] === 'float64') {
                    $qbody .= "\tif err := requireFinite(\"{$p['wire']}\", {$value}); err != nil {\n\t\treturn nil, err\n\t}\n";
                }
                if ($p['paginationMax'] !== null) {
                    $qbody .= "\tif err := checkPagination(\"{$p['wire']}\", {$value}, {$p['paginationMax']}); err != nil {\n\t\treturn nil, err\n\t}\n";
                }
                $encodedValue = $p['goType'] === 'int'
                    ? "strconv.Itoa({$value})"
                    : "strconv.FormatFloat({$value}, 'f', -1, 64)";
                if ($p['enum']) {
                    $list = implode(', ', array_map(fn($e) => '"' . $e . '"', $p['enum']));
                    $qbody .= "\tif err := checkEnum(\"{$p['wire']}\", {$encodedValue}, {$list}); err != nil {\n\t\treturn nil, err\n\t}\n";
                }
                $qbody .= "\tq.Set(\"{$p['wire']}\", {$encodedValue})\n";
            } elseif ($p['default'] !== null) {
                $qbody .= "\t{$lv} := p.{$p['field']}\n\tif {$lv} == \"\" {\n\t\t{$lv} = \"{$p['default']}\"\n\t}\n";
                $qbody .= "\tq.Set(\"{$p['wire']}\", {$lv})\n";
            } else {
                $qbody .= "\tif err := requireString(\"{$p['wire']}\", p.{$p['field']}); err != nil {\n\t\treturn nil, err\n\t}\n";
                $qbody .= "\tq.Set(\"{$p['wire']}\", p.{$p['field']})\n";
            }
        }
        foreach ($m['optional'] as $p) {
            if ($p['isFields']) {
                $qbody .= "\tif len(p.{$p['field']}) > 0 {\n\t\tq.Set(\"{$p['wire']}\", strings.Join(p.{$p['field']}, \",\"))\n\t}\n";
            } elseif ($p['isDate']) {
                $qbody .= "\tif p.{$p['field']} != nil {\n\t\tif err := requireDate(\"{$p['wire']}\", *p.{$p['field']}); err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t\tq.Set(\"{$p['wire']}\", dateString(*p.{$p['field']}))\n\t}\n";
            } else {
                $deref = '*p.' . $p['field'];
                $set = match ($p['goType']) {
                    'int'     => "strconv.Itoa({$deref})",
                    'float64' => "strconv.FormatFloat({$deref}, 'f', -1, 64)",
                    default   => $deref,
                };
                $validation = $p['goType'] === 'float64'
                    ? "\t\tif err := requireFinite(\"{$p['wire']}\", {$deref}); err != nil {\n\t\t\treturn nil, err\n\t\t}\n"
                    : '';
                if ($p['paginationMax'] !== null) {
                    $validation .= "\t\tif err := checkPagination(\"{$p['wire']}\", {$deref}, {$p['paginationMax']}); err != nil {\n\t\t\treturn nil, err\n\t\t}\n";
                }
                if ($p['enum']) {
                    $list = implode(', ', array_map(fn($e) => '"' . $e . '"', $p['enum']));
                    $qbody .= "\tif p.{$p['field']} != nil {\n{$validation}\t\tif err := checkEnum(\"{$p['wire']}\", {$set}, {$list}); err != nil {\n\t\t\treturn nil, err\n\t\t}\n";
                    $qbody .= "\t\tq.Set(\"{$p['wire']}\", {$set})\n\t}\n";
                } else {
                    $qbody .= "\tif p.{$p['field']} != nil {\n{$validation}\t\tq.Set(\"{$p['wire']}\", {$set})\n\t}\n";
                }
            }
        }

        $hasNum = false;
        foreach (array_merge($m['required'], $m['optional']) as $p) {
            if (!$p['isFields'] && !$p['isDate'] && in_array($p['goType'], ['int', 'float64'], true) && empty($p['required'])) { $hasNum = true; }
        }
        // (strconv import decided later)

        // ---- typed method ----
        $ret = match ($m['shape']) {
            'rows' => '[]map[string]interface{}',
            'strings' => '[]string',
            default => 'map[string]interface{}',
        };
        $out .= "// {$m['name']} — {$summary}.\n//\n// See {$uiURL} and\n// {$doc}\n";
        $out .= "func (s *{$type}) {$m['name']}(ctx context.Context, p {$m['params']}) ({$ret}, error) {\n";
        $out .= "\tresult, _, err := s.{$m['name']}WithResponse(ctx, p)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn result.Result, nil\n}\n\n";
        $out .= "// {$m['name']}WithResponse returns {$m['name']}'s result and sanitized HTTP metadata.\n//\n// See {$doc}\n";
        $out .= "func (s *{$type}) {$m['name']}WithResponse(ctx context.Context, p {$m['params']}) (*Result[{$ret}], *Response, error) {\n";
        $out .= str_replace('return nil, err', 'return nil, nil, err', $qbody);
        $out .= "\tresp, err := s.client.do(ctx, \"{$opName}\", \"{$path}\", q)\n\tif err != nil {\n\t\tif resp != nil {\n\t\t\treturn nil, resp.metadata, err\n\t\t}\n\t\treturn nil, nil, err\n\t}\n";
        $out .= "\tif err := validateJSONContentType(resp.header); err != nil {\n\t\treturn nil, resp.metadata, s.client.decodeError(\"{$opName}\", resp, err)\n\t}\n";
        $out .= "\tv, err := decodeResult(\"{$opName}\", resp.body)\n\tif err != nil {\n\t\treturn nil, resp.metadata, s.client.decodeError(\"{$opName}\", resp, err)\n\t}\n";
        $decode = match ($m['shape']) {
            'rows' => 'asRows',
            'strings' => 'asStrings',
            default => 'asObject',
        };
        $out .= "\tresult, err := {$decode}(\"{$opName}\", v)\n\tif err != nil {\n\t\treturn nil, resp.metadata, s.client.decodeError(\"{$opName}\", resp, err)\n\t}\n\treturn &Result[{$ret}]{Result: result}, resp.metadata, nil\n}\n\n";

        // ---- Raw variant ----
        if ($m['hasReturn']) {
            $rawbody = "\tif err := validateReturnFormat(f); err != nil {\n\t\treturn nil, err\n\t}\n\tq := url.Values{}\n\tq.Set(\"return\", string(f))\n";
            $rawbody .= substr($qbody, strlen("\tq := url.Values{}\n"));
            $out .= "// {$m['name']}Raw — {$summary}, raw download.\n//\n// Returns the response body verbatim for the chosen `return` wire\n// format (CSV or XLSX; JSON is also accepted). The typed method above\n// always requests JSON and decodes it.\n//\n// See {$uiURL} and\n// {$doc}\n";
            $out .= "func (s *{$type}) {$m['name']}Raw(ctx context.Context, p {$m['params']}, f ReturnFormat) ([]byte, error) {\n";
            $out .= $rawbody;
            $out .= "\tresp, err := s.client.do(ctx, \"{$opName}Raw\", \"{$path}\", q)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n";
            $out .= "\treturn resp.body, nil\n}\n\n";
            $out .= "// {$m['name']}Stream opens a streaming {$m['name']} download with HTTP metadata.\n// Close the returned response when finished. Success bodies are not buffered\n// or automatically retried after reading begins.\n//\n// See {$doc}\n";
            $out .= "func (s *{$type}) {$m['name']}Stream(ctx context.Context, p {$m['params']}, f ReturnFormat) (*RawResponse, error) {\n";
            $out .= $rawbody;
            $out .= "\treturn s.client.doStream(ctx, \"{$opName}Stream\", \"{$path}\", q)\n}\n\n";
        }

        $manifest[] = [
            'service' => $field, 'type' => $type, 'method' => $m['name'],
            'params' => $m['params'], 'path' => $path, 'shape' => $m['shape'],
            'required' => array_map(fn($p) => $p['wire'], $m['required']),
            'hasRaw' => $m['hasReturn'],
            'httpMethod' => 'GET', 'operationId' => $m['op']['operationId'],
            'productionURL' => $baseURL . $path,
            'documentationURL' => $uiURL, 'schemaURL' => $doc,
            'parameters' => $m['op']['parameters'] ?? [],
            'parameterFields' => (object) array_column(array_merge($m['required'], $m['optional']), 'field', 'wire'),
            'rateLimits' => $catalog[$path]['Rate Limits'] ?? null,
        ];
        $rateLimits = $catalog[$path]['Rate Limits'] ?? '';
        $hourlyLimit = str_contains($rateLimits, '100/hr when interval=1m') ? '100/hour when `interval=1m`'
            : (str_contains($rateLimits, 'Heavy 100/hr') ? '100/hour (heavy)' : ($rateLimits === '' ? 'Not listed in catalog' : '—'));
        $mdRows[] = [$tagName, $field, $m['name'], $m['params'], $m['hasReturn'], $path, $summary, $doc,
            parameterSummary($m['op']['parameters'] ?? []), $hourlyLimit, $m['op']['operationId']];
    }

    // strconv needed if any optional numeric param exists
    if (str_contains($out, 'strconv.')) {
        $out = str_replace("\t\"net/url\"\n", "\t\"net/url\"\n\t\"strconv\"\n", $out);
    }
    file_put_contents("{$pkg}/zz_gen_{$slug}.go", $out);
}

// ---- emit contract test ----
$test = "// Code generated by tools/generate.php from the CryptoDataDownload live\n// OpenAPI schema. DO NOT EDIT — regenerate instead.\n\npackage cryptodatadownload\n\nimport (\n\t\"context\"\n\t\"strings\"\n\t\"testing\"\n\t\"time\"\n)\n\n";
$test .= "// opCase is a single API operation contract test.\ntype opCase struct {\n\tname  string\n\tpath  string\n\tbody  string\n\tcall  func(*Client) error\n\textra func(*testing.T, recordedRequest)\n}\n\n";
$test .= "func TestGeneratedContractCoverage(t *testing.T) {\n\tcases := []opCase{\n";

foreach ($groups as $slug => $methods) {
    [$type, $field] = $services[$slug];
    $test .= "\t\t// --- {$field} ---\n";
    foreach ($methods as $m) {
        $example = $m['example'] ?? (object) ['result' => []];
        $fixture = json_encode($example, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR);
        // A JSON string literal is also a valid Go string literal, except for
        // escaped forward slashes (disabled above).
        $fixtureLiteral = json_encode($fixture, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR);
        $test .= "\t\t{\n\t\t\tname: \"{$m['name']}\",\n\t\t\tpath: \"{$m['path']}\",\n\t\t\tbody: {$fixtureLiteral},\n";
        $test .= "\t\t\tcall: func(c *Client) error {\n\t\t\t\t_, err := c.{$field}.{$m['name']}(context.Background(), {$m['params']}{";
        $args = [];
        foreach ($m['required'] as $p) {
            if ($p['isDate']) { $args[] = "{$p['field']}: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)"; }
            elseif ($p['enum'] && $p['goType'] === 'int') { $args[] = "{$p['field']}: " . (int) $p['enum'][0]; }
            elseif ($p['enum'] && $p['goType'] === 'float64') { $args[] = "{$p['field']}: " . (float) $p['enum'][0]; }
            elseif ($p['enum']) { $args[] = "{$p['field']}: \"{$p['enum'][0]}\""; }
            elseif ($p['goType'] === 'int') { $args[] = "{$p['field']}: 1"; }
            elseif ($p['goType'] === 'float64') { $args[] = "{$p['field']}: 1.5"; }
            else { $args[] = "{$p['field']}: \"TEST\""; }
        }
        $test .= implode(', ', $args);
        $test .= "})\n\t\t\t\treturn err\n\t\t\t},\n";
        // extra: assert required wire params present; expect == stub value
        $checks = [];
        foreach ($m['required'] as $p) {
            if ($p['isDate']) { $expect = '2024-01-01'; }
            elseif ($p['enum']) { $expect = $p['enum'][0]; }
            elseif ($p['goType'] === 'int') { $expect = '1'; }
            elseif ($p['goType'] === 'float64') { $expect = '1.5'; }
            else { $expect = 'TEST'; }
            $checks[] = "{$p['wire']}={$expect}";
        }
        if ($checks) {
            $test .= "\t\t\textra: func(t *testing.T, req recordedRequest) {\n";
            foreach ($checks as $ch) {
                $test .= "\t\t\t\tif !strings.Contains(req.Query, \"{$ch}\") {\n\t\t\t\t\tt.Errorf(\"query = %q, want {$ch}\", req.Query)\n\t\t\t\t}\n";
            }
            $test .= "\t\t\t},\n";
        }
        $test .= "\t\t},\n";
    }
}
$test .= "\t}\n\tif len(cases) != 116 {\n\t\tt.Fatalf(\"expected 116 operations, got %d\", len(cases))\n\t}\n\n";
$test .= <<<'GO'
	ts := newTestServer(t)
	c := newTestClient(t, ts)
	for _, tc := range cases {
		ts.setResponse(200, []byte(tc.body))
		if err := tc.call(c); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		req := ts.lastRequest()
		if req.Method != "GET" {
			t.Errorf("%s: method = %q, want GET", tc.name, req.Method)
		}
		if req.Path != tc.path {
			t.Errorf("%s: path = %q, want %q", tc.name, req.Path, tc.path)
		}
		requireToken(t, req)
		if tc.extra != nil {
			tc.extra(t, req)
		}
	}
	if ts.requestCount() != len(cases) {
		t.Fatalf("request count = %d, want %d", ts.requestCount(), len(cases))
	}
}
GO;
file_put_contents("{$pkg}/zz_contract_test.go", $test);

// ---- manifest + coverage doc ----
file_put_contents("{$pkg}/testdata/operations.json", json_encode($manifest, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) . "\n");

$md = "# API coverage — {$total} operations\n\n";
$md .= "Generated from the live schema ({$schemaURL}). Method names derive from each\noperation's path suffix inside its service group; `*Raw` variants exist only on\noperations whose schema declares a `return` parameter.\n\n";
$md .= "Every typed method also has a `*WithResponse` companion returning `Result[T]`\nand HTTP metadata. Each `*Raw` method also has a `*Stream` companion.\nThese variants share one operation and do not increase the 116-operation count.\n\n";
$md .= "Parameters, defaults and enums below come from the OpenAPI snapshot.\n`auth_token` is configured only through `WithQueryTokenAuth`; it is never sent by default.\nHourly limits come from the [official CSV catalog]({$catalogURL}) snapshot.\nAll catalog entries list 60 requests/minute and 10,000/day; the additional\nhourly limits below require application-level budgeting.\n\n";
$cur = null;
foreach ($mdRows as [$tag, $acc, $name, $params, $hasRaw, $path, $summary, $doc, $parameters, $hourlyLimit, $operationId]) {
    if ($tag !== $cur) {
        $cur = $tag;
        $md .= "\n## {$tag}\n\n| Method / operationId | Endpoint | Parameters | Raw / Stream | Additional hourly limit | Docs |\n|---|---|---|---|---|---|\n";
    }
    $operationId = markdownCell($operationId);
    $summary = markdownCell($summary);
    $md .= "| `{$acc}.{$name}(ctx, {$params}{})`<br>`{$operationId}` | `GET {$baseURL}{$path}` — {$summary} | {$parameters} | " . ($hasRaw ? 'yes' : '—') . " | {$hourlyLimit} | [UI]({$uiURL}) · [schema]({$doc}) |\n";
}
file_put_contents("{$pkg}/docs/api-coverage.md", $md);

// Keep the README's complete endpoint index in sync without duplicating
// parameter descriptions or editing the surrounding hand-written guide.
$readmeCoverage = "<details>\n<summary>All 116 operations and method variants</summary>\n\n";
$readmeCoverage .= "| SDK methods | Method / exact path | Production URL | Official documentation |\n|---|---|---|---|\n";
foreach ($mdRows as [$tag, $acc, $name, $params, $hasRaw, $path, $summary, $doc]) {
    $variants = ["`{$acc}.{$name}`", "`{$acc}.{$name}WithResponse`"];
    if ($hasRaw) {
        $variants[] = "`{$acc}.{$name}Raw`";
        $variants[] = "`{$acc}.{$name}Stream`";
    }
    $readmeCoverage .= '| ' . implode('<br>', $variants) . " | `GET {$schema['basePath']}{$path}` | [production]({$baseURL}{$path}) | [Swagger UI]({$uiURL}) · [OpenAPI record]({$doc}) |\n";
}
$readmeCoverage .= "\n</details>\n";
$readmePath = "{$pkg}/README.md";
if (is_file($readmePath)) {
    $readme = file_get_contents($readmePath);
    $readme = preg_replace_callback(
        '/<!-- BEGIN GENERATED ENDPOINT COVERAGE -->.*?<!-- END GENERATED ENDPOINT COVERAGE -->/s',
        fn() => "<!-- BEGIN GENERATED ENDPOINT COVERAGE -->\n{$readmeCoverage}<!-- END GENERATED ENDPOINT COVERAGE -->",
        $readme
    );
    file_put_contents($readmePath, $readme);
}

echo "generated {$total} methods across " . count($groups) . " services\n";
