// ============================================================
// ContentStudio ? Looker Studio Community Connector
// main.gs ? Auth, Config, Schema dispatch, Data dispatch, Shared utilities
// ============================================================
//
// UPDATED: Supports pre-filled config via Looker Studio deep links.
//
// TWO DEEP LINK MODES:
//
// MODE 1 ? Template Report (Linking API):
//   https://lookerstudio.google.com/reporting/create?
//     c.reportId=<TEMPLATE_REPORT_ID>
//     &ds.ds0.connector=community
//     &ds.ds0.connectorId=<DEPLOYMENT_ID>
//     &ds.ds0.access_token=<API_KEY>
//     &ds.ds0.workspace_id=<WORKSPACE_ID>
//     &ds.ds0.platform=<PLATFORM>
//     &ds.ds0.account_id=<ACCOUNT_ID>
//
// MODE 2 ? Data Source Only (Direct Link):
//   https://lookerstudio.google.com/datasources/create?
//     connectorId=<DEPLOYMENT_ID>
//     &connectorConfig=<URL_ENCODED_JSON>
//
// The original stepped flow is fully preserved as a fallback.
// ============================================================

var cc = DataStudioApp.createCommunityConnector();

// ?? Environment config ????????????????????????????????????????
function getEnv() {
    return {
        ANALYTICS:         'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
        ANALYTICS_GO:      'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
        ANALYTICS_BACKEND: 'https://qa-api.contentstudio.io'
    };
}

// ?? Auth ??????????????????????????????????????????????????????

function isAdminUser() { return true; }

function getAuthType() {
    return cc.newAuthTypeResponse()
        .setAuthType(cc.AuthType.NONE)
        .build();
}

// ?? Config ????????????????????????????????????????????????????

function getConfig(request) {
    var config    = cc.getConfig();
    var p         = request.configParams || {};
    var userProps = PropertiesService.getUserProperties();
    var env       = getEnv();

    // ?? Step 0: Persist token if provided ??????????????????????
    if (p.access_token && p.access_token.trim() !== '') {
        userProps.setProperty('cs_token', p.access_token.trim());
    }

    var token = userProps.getProperty('cs_token') || '';

    // ?? No token yet ? show input box and stop ?????????????????
    if (!token) {
        config.newTextInput()
            .setId('access_token')
            .setName('ContentStudio API Key')
            .setHelpText('Paste your ContentStudio API key and click NEXT. Generate one from Analytics ? Looker Studio in ContentStudio.');
        config.setIsSteppedConfig(true);
        config.setDateRangeRequired(true);
        return config.build();
    }

    // ?? Token exists ? show connected status ???????????????????
    config.newInfo()
        .setId('access_token')
        .setText('? Connected via ContentStudio.');

    var wsId      = p.workspace_id || '';
    var platform  = p.platform     || '';
    var accountId = p.account_id   || '';

    // ??????????????????????????????????????????????????????????????
    // FAST PATH: All params pre-filled via ContentStudio deep link.
    // Detected by the presence of workspace_name or account_name ?
    // these are only sent in the deep link URL, never by the wizard.
    // This ensures re-editing a manually configured data source still
    // goes through the stepped path and re-fetches accounts.
    // ??????????????????????????????????????????????????????????????
    if (wsId && platform && accountId && (p.workspace_name || p.account_name)) {
        var wsName  = p.workspace_name || wsId;
        var accName = p.account_name   || accountId;

        config.newSelectSingle()
            .setId('workspace_id')
            .setName('Workspace')
            .setHelpText('? Pre-selected from ContentStudio.')
            .addOption(config.newOptionBuilder().setLabel(wsName).setValue(wsId));

        config.newSelectSingle()
            .setId('platform')
            .setName('Platform')
            .setHelpText('? Pre-selected from ContentStudio.')
            .addOption(config.newOptionBuilder().setLabel(_platformDisplayName(platform)).setValue(platform));

        config.newSelectSingle()
            .setId('account_id')
            .setName('Account')
            .setHelpText('? Pre-selected. Click CONNECT to finish.')
            .addOption(config.newOptionBuilder().setLabel(accName).setValue(accountId));

        config.setIsSteppedConfig(false);
        config.setDateRangeRequired(true);
        return config.build();
    }

    // ??????????????????????????????????????????????????????????????
    // WIZARD PATH: All three selectors shown together.
    //
    // Workspace and platform are always rendered. Accounts are
    // fetched and rendered as soon as workspace + platform are known.
    // setIsSteppedConfig(true) shows NEXT (re-calls getConfig with
    // updated values); false shows CONNECT.
    // ??????????????????????????????????????????????????????????????

    _addWorkspaceSelector(config, env, token, wsId, false);
    _addPlatformSelector(config, platform);

    if (wsId && platform) {
        var fetchedAccounts = _fetchAccountsForPlatform(env, token, wsId, platform);
        var accountValid = !!(accountId && fetchedAccounts && fetchedAccounts.some(function(a) {
            return String(a.value) === String(accountId);
        }));
        _addAccountSelector(config, fetchedAccounts, platform, accountId);
        config.setIsSteppedConfig(!accountValid);
    } else {
        config.setIsSteppedConfig(true);
    }

    config.setDateRangeRequired(true);
    return config.build();
}

// ?? Config helpers ????????????????????????????????????????????

/**
 * @param {boolean} prefillMode - When true (fast path), injects the
 *   current value as the first option to guarantee selection.
 */
function _addWorkspaceSelector(config, env, token, currentWsId, prefillMode) {
    var wsSelect = config.newSelectSingle()
        .setId('workspace_id')
        .setName('Step 1: Select Workspace')
        .setHelpText(currentWsId
            ? '? Workspace selected. Change only if needed.'
            : 'Select your workspace and click NEXT.'
        );

    // In prefill mode, inject the value first to guarantee selection
    var injectedValue = false;

    try {
        var wsResult = fetchAllWorkspaces(env.ANALYTICS_BACKEND, token);
        if (wsResult.authFailed) {
            PropertiesService.getUserProperties().deleteProperty('cs_token');
            wsSelect.addOption(config.newOptionBuilder()
                .setLabel('Auth failed (HTTP ' + wsResult.code + ') ? re-enter your API key').setValue(''));
        } else if (!wsResult.data.length) {
            // No workspaces from API ? if prefilling, still add the value
            if (prefillMode && currentWsId) {
                wsSelect.addOption(config.newOptionBuilder()
                    .setLabel('Workspace ' + currentWsId).setValue(String(currentWsId)));
                injectedValue = true;
            } else {
                wsSelect.addOption(config.newOptionBuilder()
                    .setLabel('No workspaces found for this API key').setValue(''));
            }
        } else {
            // If prefilling, ensure the current value appears first
            if (prefillMode && currentWsId) {
                var found = wsResult.data.some(function(ws) {
                    return String(ws._id) === String(currentWsId);
                });
                if (!found) {
                    // Value not in API results ? inject it anyway
                    wsSelect.addOption(config.newOptionBuilder()
                        .setLabel('Workspace ' + currentWsId).setValue(String(currentWsId)));
                    injectedValue = true;
                }
            }

            wsResult.data.forEach(function(ws) {
                if (ws._id && ws.name) {
                    wsSelect.addOption(config.newOptionBuilder()
                        .setLabel(String(ws.name)).setValue(String(ws._id)));
                }
            });
        }
    } catch(e) {
        if (prefillMode && currentWsId) {
            wsSelect.addOption(config.newOptionBuilder()
                .setLabel('Workspace ' + currentWsId).setValue(String(currentWsId)));
        } else {
            wsSelect.addOption(config.newOptionBuilder()
                .setLabel('Error: ' + e.message).setValue(''));
        }
    }
}

function _platformDisplayName(platform) {
    var names = {
        instagram: 'Instagram', facebook: 'Facebook', linkedin: 'LinkedIn',
        tiktok: 'TikTok', youtube: 'YouTube', pinterest: 'Pinterest',
        twitter: 'X (Twitter)', gmb: 'Google Business Profile'
    };
    return names[platform] || platform;
}

function _addPlatformSelector(config, currentPlatform) {
    config.newSelectSingle()
        .setId('platform')
        .setName('Step 2: Select Platform')
        .setHelpText(currentPlatform
            ? '? Platform selected. Change only if needed.'
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

// Fetches accounts from the API. Returns [{label, value, validity}] or null on error.
function _fetchAccountsForPlatform(env, token, wsId, platform) {
    try {
        var allAccounts = [];
        var page = 1;
        var lastPage = 1;
        do {
            var res = UrlFetchApp.fetch(env.ANALYTICS_BACKEND + '/api/v1/workspaces/' + wsId + '/accounts?platform=' + encodeURIComponent(platform) + '&per_page=100&page=' + page, {
                method: 'get',
                headers: { 'X-API-Key': token, 'Content-Type': 'application/json' },
                muteHttpExceptions: true,
                followRedirects: true
            });
            if (res.getResponseCode() !== 200) return null;
            var body = JSON.parse(res.getContentText());
            lastPage = body.last_page || 1;
            allAccounts = allAccounts.concat(body.data || []);
            page++;
        } while (page <= lastPage);
        return getAccountsForPlatform(allAccounts, platform);
    } catch(e) {
        return null;
    }
}

// Renders the account selector from a pre-fetched accounts list.
function _addAccountSelector(config, accounts, platform, currentAccountId) {
    var accSelect = config.newSelectSingle()
        .setId('account_id')
        .setName('Step 3: Select Account')
        .setHelpText(currentAccountId
            ? '? Account selected. Click CONNECT to finish.'
            : 'Select your account and click NEXT.'
        );

    if (accounts === null) {
        accSelect.addOption(config.newOptionBuilder()
            .setLabel('Failed to load accounts ? check your API key').setValue(''));
        return;
    }
    if (!accounts.length) {
        accSelect.addOption(config.newOptionBuilder()
            .setLabel('No ' + platform + ' accounts found in this workspace').setValue(''));
        return;
    }
    accounts.forEach(function(acc) {
        var label = acc.label;
        var v = acc.validity || '';
        if (v === 'expired' || v === 'invalid') label += ' ?? (' + v + ')';
        accSelect.addOption(config.newOptionBuilder()
            .setLabel(label).setValue(String(acc.value)));
    });
}

function getAccountsForPlatform(accounts, platform) {
    // accounts is the data array from /api/v1/workspaces/{wsId}/accounts
    // Each item: { _id: platform_identifier, platform, account_name, validity, ... }
    var seen = {};
    var unique = [];
    accounts.forEach(function(acc) {
        var id = String(acc._id || '');
        if (id && !seen[id]) { seen[id] = true; unique.push(acc); }
    });
    return unique.map(function(acc) {
        return {
            label:    String(acc.account_name || acc._id || 'Unknown'),
            value:    String(acc._id || ''),
            validity: acc.validity || ''
        };
    });
}

function _getPlatformId(acc, platform) {
    switch(platform) {
        case 'facebook':  return String(acc.facebook_id           || acc.platform_identifier || acc._id || '');
        case 'instagram': return String(acc.instagram_id          || acc.platform_identifier || acc._id || '');
        case 'linkedin':  return String(acc.linkedin_id           || acc.platform_identifier || acc._id || '');
        case 'twitter':   return String(acc.twitter_id            || acc.platform_identifier || acc._id || '');
        case 'tiktok':    return String(acc.platform_identifier   || acc._id || '');
        case 'youtube':   return String(acc.platform_identifier   || acc._id || '');
        case 'pinterest': return String(acc.profile_id || acc.pinterest_id || acc.platform_identifier || acc._id || '');
        case 'gmb':       return String(acc.platform_identifier   || acc._id || '');
        default:          return String(acc.platform_identifier   || acc._id || '');
    }
}

// ?? Schema ????????????????????????????????????????????????????

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

// ?? Data ??????????????????????????????????????????????????????

function getData(request) {
    var env      = getEnv();
    var p        = request.configParams || {};
    var token    = PropertiesService.getUserProperties().getProperty('cs_token') || '';
    var platform = p.platform || 'instagram';

    if (!token)          throw new Error('API key missing. Open Analytics ? Looker Studio in ContentStudio, generate a key, and reconnect.');
    if (!p.workspace_id) throw new Error('Missing workspace_id. Please reconnect the data source.');
    if (!p.account_id)   throw new Error('Missing account_id. Please reconnect the data source.');

    // Cache timezone per workspace to avoid fetching all workspaces on every chart render.
    var tzCacheKey = 'tz_' + p.workspace_id;
    var userProps  = PropertiesService.getUserProperties();
    var tz         = userProps.getProperty(tzCacheKey) || '';
    if (!tz) {
        try {
            var wsResult = fetchAllWorkspaces(env.ANALYTICS_BACKEND, token);
            wsResult.data.forEach(function(ws) {
                if (ws._id === p.workspace_id) tz = ws.timezone || 'UTC';
            });
        } catch(e) {}
        if (!tz) tz = 'UTC';
        userProps.setProperty(tzCacheKey, tz);
    }

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

// ?? Shared HTTP helper ????????????????????????????????????????

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

// ?? Shared URL helper ?????????????????????????????????????????

function buildBaseParams(p, extra) {
    var tz = p.timezone || 'UTC';
    try { tz = decodeURIComponent(tz); } catch(e) {}
    return '?start_date='  + encodeURIComponent(p.start_date)
        + '&end_date='       + encodeURIComponent(p.end_date)
        + '&workspace_id='   + encodeURIComponent(p.workspace_id)
        + '&timezone='       + encodeURIComponent(tz)
        + (extra || '');
}

// ?? Field-routing utility ?????????????????????????????????????

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

// ?? Date utility ??????????????????????????????????????????????

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

// ?? Workspace helper ?????????????????????????????????????????

function fetchAllWorkspaces(backendBase, token) {
    var allWorkspaces = [];
    var page = 1;
    var lastPage = 1;
    var maxPages = 3; // cap to avoid bandwidth quota errors
    do {
        var res = UrlFetchApp.fetch(backendBase + '/api/v1/workspaces?per_page=100&page=' + page, {
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
    } while (page <= lastPage && page <= maxPages);
    return { authFailed: false, data: allWorkspaces };
}

// ??????????????????????????????????????????????????????????????
// DEEP LINK BUILDER
// ??????????????????????????????????????????????????????????????

function buildLookerDeepLink(opts) {
    var connectorId = opts.connectorId || 'AKfycbwEZw5klORFkztSMJj8xi8Th0pu4I7bMMkdUd2c8b3ahBOXOozhSb3HuEVUY6jYy3F9';

    // ?? MODE 1: Template report (Linking API) ??????????????????
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

    // ?? MODE 2: Data source only (Direct Link) ?????????????????
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

// ?? Debug utilities ??????????????????????????????????????????

function clearProps() {
    PropertiesService.getUserProperties().deleteAllProperties();
}

function debugProps() {
    var u = PropertiesService.getUserProperties().getProperties();
    Logger.log(JSON.stringify(u, null, 2));
}

// ?? Debug: Inspect accounts for a platform ???????????????????
// Run this to see what _id, platform, and identifiers your API
// returns for LinkedIn (or any platform). Compare with the
// account_id you're passing in the deep link.

function debugAccountsForPlatform() {
    var token      = PropertiesService.getUserProperties().getProperty('cs_token');
    var wsId       = '69006797d35c9a23ec0b2732'; // workspace from the failing deep link
    var platform   = 'linkedin';
    var targetId   = '13324048';                  // account_id from the deep link

    if (!token) {
        Logger.log('No token stored. Run the connector first to set it.');
        return;
    }

    var res = UrlFetchApp.fetch('https://qa-api.contentstudio.io/api/v1/workspaces/' + wsId + '/accounts', {
        method: 'get',
        headers: { 'X-API-Key': token, 'Content-Type': 'application/json' },
        muteHttpExceptions: true
    });
    if (res.getResponseCode() !== 200) {
        Logger.log('API error: ' + res.getResponseCode() + ' ? ' + res.getContentText().substring(0, 200));
        return;
    }
    var body = JSON.parse(res.getContentText());
    var allAccounts = body.data || [];
    Logger.log('Total accounts: ' + allAccounts.length);

    var relevant = allAccounts.filter(function(acc) {
        return acc.platform && acc.platform.toLowerCase().indexOf(platform) !== -1;
    });
    Logger.log('Found ' + relevant.length + ' ' + platform + ' accounts');
    relevant.forEach(function(acc, i) {
        Logger.log('\n--- Account ' + (i + 1) + ' ---');
        Logger.log(JSON.stringify(acc, null, 2));
        Logger.log('  _getPlatformId: "' + _getPlatformId(acc, platform) + '"');
    });

    var accounts = getAccountsForPlatform(allAccounts, platform);
    Logger.log('\ngetAccountsForPlatform returns ' + accounts.length + ' accounts:');
    accounts.forEach(function(acc) {
        var match = String(acc.value) === targetId ? ' ? MATCH' : '';
        Logger.log('  label: "' + acc.label + '"  value: "' + acc.value + '"' + match);
    });

    var found = accounts.some(function(acc) { return String(acc.value) === targetId; });
    Logger.log('\nWould "' + targetId + '" be found? ' + found);
    Logger.log('RESULT: ' + (found ? 'real name shown' : 'fallback injected'));
}

// ?? Test: Data source only ???????????????????????????????????

function testDeepLink_dataSource() {
    var url = buildLookerDeepLink({
        accessToken: 'cs_a77ec4a86bf7561be327412b5d6a2dbe11019dc0bdb9a48649e702cf053b73ad',
        workspaceId: '65604b5c7b8184eba40bf742',
        platform:    'facebook',
        accountId:   '350830594784444'
    });
    Logger.log('Data source link:\n' + url);
}

// ?? Test: With report template ???????????????????????????????

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


function debugDemographics() {
  var token = PropertiesService.getUserProperties().getProperty('cs_token');
  var url = 'https://features-analytics-pipeline.contentstudio.io/analytics/overview/linkedin/followersDemographics'
    + '?start_date=2026-03-01&end_date=2026-04-22'
    + '&workspace_id=69006797d35c9a23ec0b2732'
    + '&timezone=UTC'
    + '&linkedin_id=13324048'
    + '&type=followersDemographics';

  var res = UrlFetchApp.fetch(url, {
    method: 'get',
    headers: { 'X-API-Key': token, Accept: 'application/json' },
    muteHttpExceptions: true
  });

  Logger.log('Status: ' + res.getResponseCode());
  Logger.log('Response: ' + res.getContentText().substring(0, 2000));
}