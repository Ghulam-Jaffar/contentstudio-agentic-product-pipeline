// ============================================================
// ContentStudio → Looker Studio Community Connector
// main.gs — Auth, Config, Schema dispatch, Data dispatch, Shared utilities
// ============================================================
//
// UPDATED: Supports pre-filled config via Looker Studio deep links.
//
// TWO DEEP LINK MODES:
//
// MODE 1 — Template Report (Linking API):
//   https://lookerstudio.google.com/reporting/create?
//     c.reportId=<TEMPLATE_REPORT_ID>
//     &ds.ds0.connector=community
//     &ds.ds0.connectorId=<DEPLOYMENT_ID>
//     &ds.ds0.access_token=<API_KEY>
//     &ds.ds0.workspace_id=<WORKSPACE_ID>
//     &ds.ds0.platform=<PLATFORM>
//     &ds.ds0.account_id=<ACCOUNT_ID>
//
// MODE 2 — Data Source Only (Direct Link):
//   https://lookerstudio.google.com/datasources/create?
//     connectorId=<DEPLOYMENT_ID>
//     &connectorConfig=<URL_ENCODED_JSON>
//
// In both cases, all four params arrive in request.configParams on
// the first getConfig() call, the stepped wizard is skipped, and the
// user connects with zero manual selection.
//
// The original stepped flow is fully preserved as a fallback.
// ============================================================

var cc = DataStudioApp.createCommunityConnector();

// —— Environment config ————————————————————————————————————————
function getEnv() {
    return {
        ANALYTICS:         'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
        ANALYTICS_GO:      'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
        ANALYTICS_BACKEND: 'https://qa-api.contentstudio.io'
    };
}

// —— Auth ——————————————————————————————————————————————————————

function isAdminUser() { return true; }

function getAuthType() {
    return cc.newAuthTypeResponse()
        .setAuthType(cc.AuthType.NONE)
        .build();
}

// —— Config ————————————————————————————————————————————————————
//
// FLOW LOGIC:
//   1. If access_token arrives (from deep link OR manual entry) → persist it.
//   2. If all three IDs (workspace_id, platform, account_id) are present
//      → mark config complete, skip all steps.
//   3. Otherwise fall through the original stepped wizard.
//

function getConfig(request) {
    var config    = cc.getConfig();
    var p         = request.configParams || {};
    var userProps = PropertiesService.getUserProperties();
    var env       = getEnv();

    // ── Step 0: Persist token if provided ──────────────────────
    if (p.access_token && p.access_token.trim() !== '') {
        userProps.setProperty('cs_token', p.access_token.trim());
    }

    var token = userProps.getProperty('cs_token') || '';

    // ── No token yet → show input box and stop ─────────────────
    if (!token) {
        config.newTextInput()
            .setId('access_token')
            .setName('ContentStudio API Key')
            .setHelpText('Paste your ContentStudio API key and click NEXT. Generate one from Analytics → Looker Studio in ContentStudio.');
        config.setIsSteppedConfig(true);
        config.setDateRangeRequired(true);
        return config.build();
    }

    // ── Token exists → show connected status ───────────────────
    config.newInfo()
        .setId('access_token')
        .setText('✅ Connected via ContentStudio.');

    var wsId      = p.workspace_id || '';
    var platform  = p.platform     || '';
    var accountId = p.account_id   || '';

    // ══════════════════════════════════════════════════════════════
    // FAST PATH: All params pre-filled (deep link from ContentStudio)
    // → validate they exist, mark complete, skip the wizard entirely.
    // ══════════════════════════════════════════════════════════════
    if (wsId && platform && accountId) {

        // Still expose the selectors so Looker Studio records them in
        // the saved data-source config, but mark config as complete.
        _addWorkspaceSelector(config, env, token, wsId);
        _addPlatformSelector(config, platform);
        _addAccountSelector(config, env, token, wsId, platform, accountId);

        config.setIsSteppedConfig(false);   // ← complete!
        config.setDateRangeRequired(true);
        return config.build();
    }

    // ══════════════════════════════════════════════════════════════
    // STEPPED PATH: Original wizard (fallback when deep link params
    // are missing — e.g. user adds connector manually in Looker).
    // ══════════════════════════════════════════════════════════════

    // ── Step 1: Workspace ──────────────────────────────────────
    _addWorkspaceSelector(config, env, token, wsId);

    // ── Step 2: Platform ───────────────────────────────────────
    if (wsId) {
        _addPlatformSelector(config, platform);
    }

    // ── Step 3: Account ────────────────────────────────────────
    if (wsId && platform) {
        _addAccountSelector(config, env, token, wsId, platform, accountId);
    }

    var complete = !!(wsId && platform && accountId);
    config.setIsSteppedConfig(!complete);
    config.setDateRangeRequired(true);
    return config.build();
}

// —— Config helpers (extracted for reuse) ——————————————————————

function _addWorkspaceSelector(config, env, token, currentWsId) {
    var wsSelect = config.newSelectSingle()
        .setId('workspace_id')
        .setName('Step 1: Select Workspace')
        .setHelpText(currentWsId
            ? '✅ Workspace selected. Change only if needed.'
            : 'Select your workspace and click NEXT.'
        );
    try {
        var wsResult = fetchAllWorkspaces(env.ANALYTICS_BACKEND, token);
        if (wsResult.authFailed) {
            PropertiesService.getUserProperties().deleteProperty('cs_token');
            wsSelect.addOption(config.newOptionBuilder()
                .setLabel('Auth failed (HTTP ' + wsResult.code + ') — re-enter your API key').setValue(''));
        } else if (!wsResult.data.length) {
            wsSelect.addOption(config.newOptionBuilder()
                .setLabel('No workspaces found for this API key').setValue(''));
        } else {
            wsResult.data.forEach(function(ws) {
                if (ws._id && ws.name) {
                    wsSelect.addOption(config.newOptionBuilder()
                        .setLabel(String(ws.name)).setValue(String(ws._id)));
                }
            });
        }
    } catch(e) {
        wsSelect.addOption(config.newOptionBuilder()
            .setLabel('Error: ' + e.message).setValue(''));
    }
}

function _addPlatformSelector(config, currentPlatform) {
    config.newSelectSingle()
        .setId('platform')
        .setName('Step 2: Select Platform')
        .setHelpText(currentPlatform
            ? '✅ Platform selected. Change only if needed.'
            : 'Select the platform and click NEXT.'
        )
        .addOption(config.newOptionBuilder().setLabel('Instagram').setValue('instagram'))
        .addOption(config.newOptionBuilder().setLabel('Facebook').setValue('facebook'))
        .addOption(config.newOptionBuilder().setLabel('LinkedIn').setValue('linkedin'))
        .addOption(config.newOptionBuilder().setLabel('TikTok').setValue('tiktok'))
        .addOption(config.newOptionBuilder().setLabel('YouTube').setValue('youtube'))
        .addOption(config.newOptionBuilder().setLabel('Pinterest').setValue('pinterest'))
        .addOption(config.newOptionBuilder().setLabel('X (Twitter)').setValue('twitter'))
        .addOption(config.newOptionBuilder().setLabel('Google Business Profile').setValue('gmb'));
}

function _addAccountSelector(config, env, token, wsId, platform, currentAccountId) {
    var accSelect = config.newSelectSingle()
        .setId('account_id')
        .setName('Step 3: Select Account')
        .setHelpText(currentAccountId
            ? '✅ Account selected. Click CONNECT to finish.'
            : 'Select your account and click NEXT.'
        );
    try {
        var allAccounts = [];
        var page = 1;
        var lastPage = 1;
        do {
            var accRes = UrlFetchApp.fetch(env.ANALYTICS_BACKEND + '/api/v1/workspaces/' + wsId + '/accounts?page=' + page, {
                method: 'get',
                headers: { 'X-API-Key': token, 'Content-Type': 'application/json' },
                muteHttpExceptions: true,
                followRedirects: true
            });
            var accCode = accRes.getResponseCode();
            if (accCode !== 200) {
                accSelect.addOption(config.newOptionBuilder()
                    .setLabel('Auth failed (HTTP ' + accCode + ') — re-enter your API key').setValue(''));
                break;
            }
            var accBody = JSON.parse(accRes.getContentText());
            lastPage = accBody.last_page || 1;
            allAccounts = allAccounts.concat(accBody.data || []);
            page++;
        } while (page <= lastPage);

        var accounts = getAccountsForPlatform({ data: allAccounts }, platform);
        if (!accounts.length) {
            accSelect.addOption(config.newOptionBuilder()
                .setLabel('No ' + platform + ' accounts found in this workspace').setValue(''));
        } else {
            accounts.forEach(function(acc) {
                var label = acc.label;
                var v = acc.validity || '';
                if (v === 'expired' || v === 'invalid') label += ' ⚠️ (' + v + ')';
                accSelect.addOption(config.newOptionBuilder()
                    .setLabel(label).setValue(String(acc.value)));
            });
        }
    } catch(e) {
        accSelect.addOption(config.newOptionBuilder()
            .setLabel('Error: ' + e.message).setValue(''));
    }
}

function getAccountsForPlatform(body, platform) {
    var all = body.data || [];
    var filtered = all.filter(function(acc) {
        return acc.platform === platform;
    });
    return filtered.map(function(acc) {
        return {
            label:    String(acc.account_name || acc._id || 'Unknown'),
            value:    String(acc._id || ''),
            validity: acc.validity || ''
        };
    });
}

// —— Schema ————————————————————————————————————————————————————

function getSchema(request) {
    var platform = (request.configParams || {}).platform || 'instagram';
    return { schema: buildFields(platform).build() };
}

function buildFields(platform) {
    var fields = cc.getFields();
    var types  = cc.FieldType;
    switch(platform) {
        case 'facebook':  return getFields_facebook(fields, types);
        case 'linkedin':  return getFields_linkedin(fields, types);
        case 'tiktok':    return getFields_tiktok(fields, types);
        case 'youtube':   return getFields_youtube(fields, types);
        case 'pinterest': return getFields_pinterest(fields, types);
        case 'twitter':   return getFields_twitter(fields, types);
        case 'gmb':       return getFields_gmb(fields, types);
        default:          return getFields_instagram(fields, types);
    }
}

// —— Data ——————————————————————————————————————————————————————

function getData(request) {
    var env      = getEnv();
    var p        = request.configParams || {};
    var token    = PropertiesService.getUserProperties().getProperty('cs_token') || '';
    var platform = p.platform || 'instagram';

    if (!token)          throw new Error('API key missing. Open Analytics → Looker Studio in ContentStudio, generate a key, and reconnect.');
    if (!p.workspace_id) throw new Error('Missing workspace_id. Please reconnect the data source.');
    if (!p.account_id)   throw new Error('Missing account_id. Please reconnect the data source.');

    var tz = 'UTC';
    try {
        var wsResult = fetchAllWorkspaces(env.ANALYTICS_BACKEND, token);
        wsResult.data.forEach(function(ws) {
            if (ws._id === p.workspace_id) tz = ws.timezone || 'UTC';
        });
    } catch(e) {}

    p.access_token = token;
    p.timezone     = tz;
    p.analytics    = env.ANALYTICS;
    p.analytics_go = env.ANALYTICS_GO;

    var dr       = request.dateRange || {};
    var fallback = getLast30DaysRange();
    p.start_date = dr.startDate || fallback.start_date;
    p.end_date   = dr.endDate   || fallback.end_date;

    p._reqIds = (request.fields || []).map(function(f) { return f.name; });

    var fields = buildFields(platform).forIds(p._reqIds);
    var schema = fields.asArray();
    var rows;

    try {
        switch(platform) {
            case 'facebook':  rows = getData_facebook(p);  break;
            case 'linkedin':  rows = getData_linkedin(p);  break;
            case 'tiktok':    rows = getData_tiktok(p);    break;
            case 'youtube':   rows = getData_youtube(p);   break;
            case 'pinterest': rows = getData_pinterest(p); break;
            case 'twitter':   rows = getData_twitter(p);   break;
            case 'gmb':       rows = getData_gmb(p);       break;
            default:          rows = getData_instagram(p); break;
        }
    } catch(e) {
        throw new Error('[' + platform + '] ' + e.message);
    }

    return {
        schema: fields.build(),
        rows: rows.map(function(row) {
            return {
                values: schema.map(function(f) {
                    var v = row[f.getId()];
                    if (v === undefined || v === null) {
                        return (f.getType() === cc.FieldType.YEAR_MONTH_DAY
                            || f.getType() === cc.FieldType.TEXT
                            || f.getType() === cc.FieldType.URL) ? '' : 0;
                    }
                    return v;
                })
            };
        })
    };
}

// —— Shared HTTP helper ————————————————————————————————————————

function analyticsGet(url, token) {
    var res = UrlFetchApp.fetch(url, {
        method: 'get',
        headers: { 'X-API-Key': token, Accept: 'application/json' },
        muteHttpExceptions: true
    });
    if (res.getResponseCode() !== 200) {
        throw new Error('API error (' + res.getResponseCode() + '): '
            + res.getContentText().substring(0, 300));
    }
    return JSON.parse(res.getContentText());
}

// —— Shared URL helper —————————————————————————————————————————

function buildBaseParams(p, extra) {
    var tz = p.timezone || 'UTC';
    try { tz = decodeURIComponent(tz); } catch(e) {}
    return '?start_date='  + encodeURIComponent(p.start_date)
        + '&end_date='       + encodeURIComponent(p.end_date)
        + '&workspace_id='   + encodeURIComponent(p.workspace_id)
        + '&timezone='       + encodeURIComponent(tz)
        + (extra || '');
}

// —— Field-routing utility —————————————————————————————————————

function bestMatch(reqIds, fieldGroups) {
    var best = Object.keys(fieldGroups)[0];
    var bestScore = 0;
    Object.keys(fieldGroups).forEach(function(k) {
        var score = fieldGroups[k].filter(function(id) {
            return reqIds.indexOf(id) !== -1;
        }).length;
        if (score > bestScore) { bestScore = score; best = k; }
    });
    return best;
}

// —— Date utility ——————————————————————————————————————————————

function toDateStr(iso) {
    return (iso || '').split('T')[0].replace(/-/g, '');
}

function getLast30DaysRange() {
    var now   = new Date();
    var start = new Date(now);
    start.setDate(start.getDate() - 30);

    function fmt(d) {
        var mm = String(d.getMonth() + 1).padStart(2, '0');
        var dd = String(d.getDate()).padStart(2, '0');
        return d.getFullYear() + '-' + mm + '-' + dd;
    }

    return { start_date: fmt(start), end_date: fmt(now) };
}

// —— Workspace helper —————————————————————————————————————————

function fetchAllWorkspaces(backendBase, token) {
    var allWorkspaces = [];
    var page = 1;
    var lastPage = 1;
    do {
        var res = UrlFetchApp.fetch(backendBase + '/api/v1/workspaces?page=' + page, {
            method: 'get',
            headers: { 'X-API-Key': token, 'Content-Type': 'application/json' },
            muteHttpExceptions: true,
            followRedirects: true
        });
        var code = res.getResponseCode();
        if (code !== 200) {
            return { authFailed: true, code: code, data: [] };
        }
        var body = JSON.parse(res.getContentText());
        lastPage = body.last_page || 1;
        allWorkspaces = allWorkspaces.concat(body.data || []);
        page++;
    } while (page <= lastPage);
    return { authFailed: false, data: allWorkspaces };
}

// ══════════════════════════════════════════════════════════════
// DEEP LINK BUILDER
// ══════════════════════════════════════════════════════════════
//
// TWO MODES supported:
//
// MODE 1 — With templateId (Linking API):
//   Opens a report template with pre-configured data source.
//   URL: /reporting/create
//     + c.reportId=<template>
//     + ds.<alias>.connector=community        ← REQUIRED for community connectors
//     + ds.<alias>.connectorId=<deployment>   ← your actual deployment ID
//     + ds.<alias>.<param>=<value>            ← config params individually
//
// MODE 2 — Without templateId (Direct Link):
//   Opens the data source config page with pre-filled params.
//   URL: /datasources/create
//     + connectorId=<deployment>
//     + connectorConfig=<URL-encoded JSON of all config params>
//
// Example usage from ContentStudio frontend (JavaScript):
//
//   // With template (recommended for production):
//   const url = buildLookerDeepLink({
//     templateId:  'ff026271-696a-4bf2-8140-29115808d46e',
//     accessToken: 'cs_abc123...',
//     workspaceId: '6621f1e...',
//     platform:    'facebook',
//     accountId:   '350830594784444',
//   });
//   window.open(url, '_blank');
//
//   // Without template (data source only):
//   const url = buildLookerDeepLink({
//     accessToken: 'cs_abc123...',
//     workspaceId: '6621f1e...',
//     platform:    'facebook',
//     accountId:   '350830594784444',
//   });
//   window.open(url, '_blank');
//
// ══════════════════════════════════════════════════════════════

function buildLookerDeepLink(opts) {
    var connectorId = opts.connectorId || 'AKfycbwEZw5klORFkztSMJj8xi8Th0pu4I7bMMkdUd2c8b3ahBOXOozhSb3HuEVUY6jYy3F9';

    // ── MODE 1: Template report (Linking API) ──────────────────
    if (opts.templateId) {
        var dsAlias = opts.dsAlias || 'ds0';
        return 'https://lookerstudio.google.com/reporting/create'
            + '?c.reportId='                       + encodeURIComponent(opts.templateId)
            + '&ds.' + dsAlias + '.connector=community'
            + '&ds.' + dsAlias + '.connectorId='   + encodeURIComponent(connectorId)
            + '&ds.' + dsAlias + '.access_token='  + encodeURIComponent(opts.accessToken || '')
            + '&ds.' + dsAlias + '.workspace_id='  + encodeURIComponent(opts.workspaceId || '')
            + '&ds.' + dsAlias + '.platform='      + encodeURIComponent(opts.platform    || '')
            + '&ds.' + dsAlias + '.account_id='    + encodeURIComponent(opts.accountId   || '');
    }

    // ── MODE 2: Data source only (Direct Link) ─────────────────
    var configObj = {
        access_token: opts.accessToken || '',
        workspace_id: opts.workspaceId || '',
        platform:     opts.platform    || '',
        account_id:   opts.accountId   || ''
    };
    return 'https://lookerstudio.google.com/datasources/create'
        + '?connectorId='     + encodeURIComponent(connectorId)
        + '&connectorConfig=' + encodeURIComponent(JSON.stringify(configObj));
}

// —— Debug utilities (remove before production) ————————————————

function clearProps() {
    PropertiesService.getUserProperties().deleteAllProperties();
}

function debugProps() {
    var u = PropertiesService.getUserProperties().getProperties();
    Logger.log(JSON.stringify(u, null, 2));
}

// —— Test: Data source only (no template) —————————————————————

function testDeepLink_dataSource() {
    var url = buildLookerDeepLink({
        accessToken: 'cs_a77ec4a86bf7561be327412b5d6a2dbe11019dc0bdb9a48649e702cf053b73ad',
        workspaceId: '65604b5c7b8184eba40bf742',
        platform:    'facebook',
        accountId:   '350830594784444'
    });
    Logger.log('Data source link:\n' + url);
}

// —— Test: With report template ———————————————————————————————

function testDeepLink_template() {
    var url = buildLookerDeepLink({
        templateId:  'ff026271-696a-4bf2-8140-29115808d46e',
        accessToken: 'cs_a77ec4a86bf7561be327412b5d6a2dbe11019dc0bdb9a48649e702cf053b73ad',
        workspaceId: '65604b5c7b8184eba40bf742',
        platform:    'facebook',
        accountId:   '350830594784444'
    });
    Logger.log('Template link:\n' + url);
}