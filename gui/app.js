// Copyright (C) 2014 Jakob Borg and other contributors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

/*jslint browser: true, continue: true, plusplus: true */
/*global $: false, angular: false */

'use strict';

var syncthing = angular.module('syncthing', []);
var urlbase = 'rest';

syncthing.controller('SyncthingCtrl', function ($scope, $http) {
    var prevDate = 0;
    var getOK = true;
    var restarting = false;

    $scope.connections = {};
    $scope.config = {};
    $scope.myID = '';
    $scope.nodes = [];
    $scope.configInSync = true;
    $scope.protocolChanged = false;
    $scope.errors = [];
    $scope.seenError = '';
    $scope.model = {};
    $scope.repos = {};

    // Strings before bools look better
    $scope.settings = [
    {id: 'ListenStr', descr: 'Sync Protocol Listen Addresses', type: 'text', restart: true},
    {id: 'MaxSendKbps', descr: 'Outgoing Rate Limit (KiB/s)', type: 'number', restart: true},
    {id: 'RescanIntervalS', descr: 'Rescan Interval (s)', type: 'number', restart: true},
    {id: 'ReconnectIntervalS', descr: 'Reconnect Interval (s)', type: 'number', restart: true},
    {id: 'ParallelRequests', descr: 'Max Outstanding Requests', type: 'number', restart: true},
    {id: 'MaxChangeKbps', descr: 'Max File Change Rate (KiB/s)', type: 'number', restart: true},

    {id: 'GlobalAnnEnabled', descr: 'Global Discovery', type: 'bool', restart: true},
    {id: 'LocalAnnEnabled', descr: 'Local Discovery', type: 'bool', restart: true},
    {id: 'LocalAnnPort', descr: 'Local Discovery Port', type: 'number', restart: true},
    {id: 'StartBrowser', descr: 'Start Browser', type: 'bool'},
    {id: 'UPnPEnabled', descr: 'Enable UPnP', type: 'bool'},
    ];

    $scope.guiSettings = [
    {id: 'Address', descr: 'GUI Listen Addresses', type: 'text', restart: true},
    {id: 'User', descr: 'GUI Authentication User', type: 'text', restart: true},
    {id: 'Password', descr: 'GUI Authentication Password', type: 'password', restart: true},
    {id: 'UseTLS', descr: 'Use HTTPS for GUI', type: 'bool', restart: true},
    ];

    function getSucceeded() {
        if (!getOK) {
            $scope.init();
            $('#networkError').modal('hide');
            getOK = true;
        }
        if (restarting) {
            $scope.init();
            $('#restarting').modal('hide');
            $('#shutdown').modal('hide');
            restarting = false;
        }
    }

    function getFailed() {
        if (restarting) {
            return;
        }
        if (getOK) {
            $('#networkError').modal({backdrop: 'static', keyboard: false});
            getOK = false;
        }
    }

    $scope.refresh = function () {
        $http.get(urlbase + '/system').success(function (data) {
            getSucceeded();
            $scope.system = data;
        }).error(function () {
            getFailed();
        });
        Object.keys($scope.repos).forEach(function (id) {
            $http.get(urlbase + '/model?repo=' + encodeURIComponent(id)).success(function (data) {
                $scope.model[id] = data;
            });
        });
        $http.get(urlbase + '/connections').success(function (data) {
            var now = Date.now(),
            td = (now - prevDate) / 1000,
            id;

            prevDate = now;
            for (id in data) {
                if (!data.hasOwnProperty(id)) {
                    continue;
                }
                try {
                    data[id].inbps = Math.max(0, 8 * (data[id].InBytesTotal - $scope.connections[id].InBytesTotal) / td);
                    data[id].outbps = Math.max(0, 8 * (data[id].OutBytesTotal - $scope.connections[id].OutBytesTotal) / td);
                } catch (e) {
                    data[id].inbps = 0;
                    data[id].outbps = 0;
                }
            }
            $scope.connections = data;
        });
        $http.get(urlbase + '/errors').success(function (data) {
            $scope.errors = data;
        });
    };

    $scope.repoStatus = function (repo) {
        if (typeof $scope.model[repo] === 'undefined') {
            return 'Unknown';
        }

        if ($scope.model[repo].invalid !== '') {
            return 'Stopped';
        }

        var state = '' + $scope.model[repo].state;
        state = state[0].toUpperCase() + state.substr(1);

        if (state == "Syncing" || state == "Idle") {
            state += " (" + $scope.syncPercentage(repo) + "%)";
        }

        return state;
    };

    $scope.repoClass = function (repo) {
        if (typeof $scope.model[repo] === 'undefined') {
            return 'info';
        }

        if ($scope.model[repo].invalid !== '') {
            return 'danger';
        }

        var state = '' + $scope.model[repo].state;
        if (state == 'idle') {
            return 'success';
        }
        if (state == 'syncing') {
            return 'primary';
        }
        return 'info';
    };

    $scope.syncPercentage = function (repo) {
        if (typeof $scope.model[repo] === 'undefined') {
            return 100;
        }
        if ($scope.model[repo].globalBytes === 0) {
            return 100;
        }

        var pct = 100 * $scope.model[repo].inSyncBytes / $scope.model[repo].globalBytes;
        return Math.floor(pct);
    };

    $scope.nodeStatus = function (nodeCfg) {
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            if (conn.Completion === 100) {
                return 'Up to Date';
            } else {
                return 'Syncing (' + conn.Completion + '%)';
            }
        }

        return 'Disconnected';
    };

    $scope.nodeIcon = function (nodeCfg) {
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            if (conn.Completion === 100) {
                return 'ok';
            } else {
                return 'refresh';
            }
        }

        return 'minus';
    };

    $scope.nodeClass = function (nodeCfg) {
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            if (conn.Completion === 100) {
                return 'success';
            } else {
                return 'primary';
            }
        }

        return 'info';
    };

    $scope.nodeAddr = function (nodeCfg) {
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            return conn.Address;
        }
        return '?';
    };

    $scope.nodeCompletion = function (nodeCfg) {
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            return conn.Completion + '%';
        }
        return '';
    };

    $scope.nodeVer = function (nodeCfg) {
        if (nodeCfg.NodeID === $scope.myID) {
            return $scope.version;
        }
        var conn = $scope.connections[nodeCfg.NodeID];
        if (conn) {
            return conn.ClientVersion;
        }
        return '?';
    };

    $scope.findNode = function (nodeID) {
        var matches = $scope.nodes.filter(function (n) { return n.NodeID == nodeID; });
        if (matches.length != 1) {
            return undefined;
        }
        return matches[0];
    };

    $scope.nodeName = function (nodeCfg) {
        if (typeof nodeCfg === 'undefined') {
            return "";
        }
        if (nodeCfg.Name) {
            return nodeCfg.Name;
        }
        return nodeCfg.NodeID.substr(0, 6);
    };

    $scope.thisNodeName = function () {
        var node = $scope.thisNode();
        if (typeof node === 'undefined') {
            return "(unknown node)";
        }
        if (node.Name) {
            return node.Name;
        }
        return node.NodeID.substr(0, 6);
    };

    $scope.editSettings = function () {
        // Make a working copy
        $scope.config.workingOptions = angular.copy($scope.config.Options);
        $scope.config.workingGUI = angular.copy($scope.config.GUI);
        $('#settings').modal({backdrop: 'static', keyboard: true});
    };

    $scope.saveSettings = function () {
        // Make sure something changed
        var changed = ! angular.equals($scope.config.Options, $scope.config.workingOptions) ||
                      ! angular.equals($scope.config.GUI, $scope.config.workingGUI);
        if(changed){
            // see if protocol will need to be changed on restart
            if($scope.config.GUI.UseTLS !== $scope.config.workingGUI.UseTLS){
                $scope.protocolChanged = true;
            }

            // Apply new settings locally
            $scope.config.Options = angular.copy($scope.config.workingOptions);
            $scope.config.GUI = angular.copy($scope.config.workingGUI);

            $scope.configInSync = false;
            $scope.config.Options.ListenAddress = $scope.config.Options.ListenStr.split(',').map(function (x) { return x.trim(); });
            $http.post(urlbase + '/config', JSON.stringify($scope.config), {headers: {'Content-Type': 'application/json'}});
        }

        $('#settings').modal("hide");
    };

    $scope.restart = function () {
        restarting = true;
        $('#restarting').modal({backdrop: 'static', keyboard: false});
        $http.post(urlbase + '/restart');
        $scope.configInSync = true;

        // Switch webpage protocol if needed
        if($scope.protocolChanged){
            var protocol = 'http';

            if($scope.config.GUI.UseTLS){
               protocol = 'https';
            }

            setTimeout(function(){
                window.location.protocol = protocol;
            }, 1000);

            $scope.protocolChanged = false;
        }
    };

    $scope.shutdown = function () {
        restarting = true;
        $http.post(urlbase + '/shutdown').success(function () {
            $('#shutdown').modal({backdrop: 'static', keyboard: false});
        });
        $scope.configInSync = true;
    };

    $scope.editNode = function (nodeCfg) {
        $scope.currentNode = $.extend({}, nodeCfg);
        $scope.editingExisting = true;
        $scope.editingSelf = (nodeCfg.NodeID == $scope.myID);
        $scope.currentNode.AddressesStr = nodeCfg.Addresses.join(', ');
        $scope.nodeEditor.$setPristine();
        $('#editNode').modal({backdrop: 'static', keyboard: true});
    };

    $scope.idNode = function () {
        $('#idqr').modal('show');
    };

    $scope.addNode = function () {
        $scope.currentNode = {AddressesStr: 'dynamic'};
        $scope.editingExisting = false;
        $scope.editingSelf = false;
        $scope.nodeEditor.$setPristine();
        $('#editNode').modal({backdrop: 'static', keyboard: true});
    };

    $scope.deleteNode = function () {
        $('#editNode').modal('hide');
        if (!$scope.editingExisting) {
            return;
        }

        $scope.nodes = $scope.nodes.filter(function (n) {
            return n.NodeID !== $scope.currentNode.NodeID;
        });
        $scope.config.Nodes = $scope.nodes;

        for (var id in $scope.repos) {
            $scope.repos[id].Nodes = $scope.repos[id].Nodes.filter(function (n) {
                return n.NodeID !== $scope.currentNode.NodeID;
            });
        }

        $scope.configInSync = false;
        $http.post(urlbase + '/config', JSON.stringify($scope.config), {headers: {'Content-Type': 'application/json'}});
    };

    $scope.saveNode = function () {
        var nodeCfg, done, i;

        $scope.configInSync = false;
        $('#editNode').modal('hide');
        nodeCfg = $scope.currentNode;
        nodeCfg.NodeID = nodeCfg.NodeID.replace(/ /g, '').replace(/-/g, '').toUpperCase().trim();
        nodeCfg.Addresses = nodeCfg.AddressesStr.split(',').map(function (x) { return x.trim(); });

        done = false;
        for (i = 0; i < $scope.nodes.length; i++) {
            if ($scope.nodes[i].NodeID === nodeCfg.NodeID) {
                $scope.nodes[i] = nodeCfg;
                done = true;
                break;
            }
        }

        if (!done) {
            $scope.nodes.push(nodeCfg);
        }

        $scope.nodes.sort(nodeCompare);
        $scope.config.Nodes = $scope.nodes;

        $http.post(urlbase + '/config', JSON.stringify($scope.config), {headers: {'Content-Type': 'application/json'}});
    };

    $scope.otherNodes = function () {
        return $scope.nodes.filter(function (n){
            return n.NodeID !== $scope.myID;
        });
    };

    $scope.thisNode = function () {
        var i, n;

        for (i = 0; i < $scope.nodes.length; i++) {
            n = $scope.nodes[i];
            if (n.NodeID === $scope.myID) {
                return n;
            }
        }
    };

    $scope.allNodes = function () {
        var nodes = $scope.otherNodes();
        nodes.push($scope.thisNode());
        return nodes;
    };

    $scope.errorList = function () {
        return $scope.errors.filter(function (e) {
            return e.Time > $scope.seenError;
        });
    };

    $scope.clearErrors = function () {
        $scope.seenError = $scope.errors[$scope.errors.length - 1].Time;
        $http.post(urlbase + '/error/clear');
    };

    $scope.friendlyNodes = function (str) {
        for (var i = 0; i < $scope.nodes.length; i++) {
            var cfg = $scope.nodes[i];
            str = str.replace(cfg.NodeID, $scope.nodeName(cfg));
        }
        return str;
    };

    $scope.repoList = function () {
        return repoList($scope.repos);
    };

    $scope.editRepo = function (nodeCfg) {
        $scope.currentRepo = angular.copy(nodeCfg);
        $scope.currentRepo.selectedNodes = {};
        $scope.currentRepo.Nodes.forEach(function (n) {
            $scope.currentRepo.selectedNodes[n.NodeID] = true;
        });
        if ($scope.currentRepo.Versioning && $scope.currentRepo.Versioning.Type === "simple") {
            $scope.currentRepo.simpleFileVersioning = true;
            $scope.currentRepo.simpleKeep = +$scope.currentRepo.Versioning.Params.keep;
        }
        $scope.currentRepo.simpleKeep = $scope.currentRepo.simpleKeep || 5;
        $scope.editingExisting = true;
        $scope.repoEditor.$setPristine();
        $('#editRepo').modal({backdrop: 'static', keyboard: true});
    };

    $scope.addRepo = function () {
        $scope.currentRepo = {selectedNodes: {}};
        $scope.editingExisting = false;
        $scope.repoEditor.$setPristine();
        $('#editRepo').modal({backdrop: 'static', keyboard: true});
    };

    $scope.saveRepo = function () {
        var repoCfg, done, i;

        $scope.configInSync = false;
        $('#editRepo').modal('hide');
        repoCfg = $scope.currentRepo;
        repoCfg.Nodes = [];
        repoCfg.selectedNodes[$scope.myID] = true;
        for (var nodeID in repoCfg.selectedNodes) {
            if (repoCfg.selectedNodes[nodeID] === true) {
                repoCfg.Nodes.push({NodeID: nodeID});
            }
        }
        delete repoCfg.selectedNodes;

        if (repoCfg.simpleFileVersioning) {
            repoCfg.Versioning = {
                'Type': 'simple',
                'Params': {
                    'keep': '' + repoCfg.simpleKeep,
                }
            };
            delete repoCfg.simpleFileVersioning;
            delete repoCfg.simpleKeep;
        } else {
            delete repoCfg.Versioning;
        }

        $scope.repos[repoCfg.ID] = repoCfg;
        $scope.config.Repositories = repoList($scope.repos);

        $http.post(urlbase + '/config', JSON.stringify($scope.config), {headers: {'Content-Type': 'application/json'}});
    };

    $scope.sharesRepo = function(repoCfg) {
        var names = [];
        repoCfg.Nodes.forEach(function (node) {
            names.push($scope.nodeName($scope.findNode(node.NodeID)));
        });
        names.sort();
        return names.join(", ");
    };

    $scope.deleteRepo = function () {
        $('#editRepo').modal('hide');
        if (!$scope.editingExisting) {
            return;
        }

        delete $scope.repos[$scope.currentRepo.ID];
        $scope.config.Repositories = repoList($scope.repos);

        $scope.configInSync = false;
        $http.post(urlbase + '/config', JSON.stringify($scope.config), {headers: {'Content-Type': 'application/json'}});
    };

    $scope.init = function() {
        $http.get(urlbase + '/version').success(function (data) {
            $scope.version = data;
        });

        $http.get(urlbase + '/system').success(function (data) {
            $scope.system = data;
            $scope.myID = data.myID;
        });

        $http.get(urlbase + '/config').success(function (data) {
            $scope.config = data;
            $scope.config.Options.ListenStr = $scope.config.Options.ListenAddress.join(', ');

            $scope.nodes = $scope.config.Nodes;
            $scope.nodes.sort(nodeCompare);

            $scope.repos = repoMap($scope.config.Repositories);

            $scope.refresh();
        });

        $http.get(urlbase + '/config/sync').success(function (data) {
            $scope.configInSync = data.configInSync;
        });
    };

    $scope.init();
    setInterval($scope.refresh, 10000);
});

function nodeCompare(a, b) {
    if (typeof a.Name !== 'undefined' && typeof b.Name !== 'undefined') {
        if (a.Name < b.Name)
            return -1;
        return a.Name > b.Name;
    }
    if (a.NodeID < b.NodeID) {
        return -1;
    }
    return a.NodeID > b.NodeID;
}

function repoCompare(a, b) {
    if (a.Directory < b.Directory) {
        return -1;
    }
    return a.Directory > b.Directory;
}

function repoMap(l) {
    var m = {};
    l.forEach(function (r) {
        m[r.ID] = r;
    });
    return m;
}

function repoList(m) {
    var l = [];
    for (var id in m) {
        l.push(m[id]);
    }
    l.sort(repoCompare);
    return l;
}

function decimals(val, num) {
    var digits, decs;

    if (val === 0) {
        return 0;
    }

    digits = Math.floor(Math.log(Math.abs(val)) / Math.log(10));
    decs = Math.max(0, num - digits);
    return decs;
}

syncthing.filter('natural', function () {
    return function (input, valid) {
        return input.toFixed(decimals(input, valid));
    };
});

syncthing.filter('binary', function () {
    return function (input) {
        if (input === undefined) {
            return '0 ';
        }
        if (input > 1024 * 1024 * 1024) {
            input /= 1024 * 1024 * 1024;
            return input.toFixed(decimals(input, 2)) + ' Gi';
        }
        if (input > 1024 * 1024) {
            input /= 1024 * 1024;
            return input.toFixed(decimals(input, 2)) + ' Mi';
        }
        if (input > 1024) {
            input /= 1024;
            return input.toFixed(decimals(input, 2)) + ' Ki';
        }
        return Math.round(input) + ' ';
    };
});

syncthing.filter('metric', function () {
    return function (input) {
        if (input === undefined) {
            return '0 ';
        }
        if (input > 1000 * 1000 * 1000) {
            input /= 1000 * 1000 * 1000;
            return input.toFixed(decimals(input, 2)) + ' G';
        }
        if (input > 1000 * 1000) {
            input /= 1000 * 1000;
            return input.toFixed(decimals(input, 2)) + ' M';
        }
        if (input > 1000) {
            input /= 1000;
            return input.toFixed(decimals(input, 2)) + ' k';
        }
        return Math.round(input) + ' ';
    };
});

syncthing.filter('short', function () {
    return function (input) {
        return input.substr(0, 6);
    };
});

syncthing.filter('alwaysNumber', function () {
    return function (input) {
        if (input === undefined) {
            return 0;
        }
        return input;
    };
});

syncthing.filter('chunkID', function () {
    return function (input) {
        if (input === undefined)
            return "";
        var parts = input.match(/.{1,6}/g);
        if (!parts)
            return "";
        return parts.join('-');
    };
});

syncthing.filter('shortPath', function () {
    return function (input) {
        if (input === undefined)
            return "";
        var parts = input.split(/[\/\\]/);
        if (!parts || parts.length <= 3) {
            return input;
        }
        return ".../" + parts.slice(parts.length-2).join("/");
    };
});

syncthing.filter('clean', function () {
    return function (input) {
        return encodeURIComponent(input).replace(/%/g, '').replace(/\./g, '_');
    };
});

syncthing.directive('optionEditor', function () {
    return {
        restrict: 'C',
        replace: true,
        transclude: true,
        scope: {
            setting: '=setting',
        },
        template: '<input type="text" ng-model="config.Options[setting.id]"></input>',
    };
});

syncthing.directive('uniqueRepo', function() {
    return {
        require: 'ngModel',
        link: function(scope, elm, attrs, ctrl) {
            ctrl.$parsers.unshift(function(viewValue) {
                if (scope.editingExisting) {
                    // we shouldn't validate
                    ctrl.$setValidity('uniqueRepo', true);
                } else if (scope.repos[viewValue]) {
                    // the repo exists already
                    ctrl.$setValidity('uniqueRepo', false);
                } else {
                    // the repo is unique
                    ctrl.$setValidity('uniqueRepo', true);
                }
                return viewValue;
            });
        }
    };
});

syncthing.directive('validNodeid', function() {
    return {
        require: 'ngModel',
        link: function(scope, elm, attrs, ctrl) {
            ctrl.$parsers.unshift(function(viewValue) {
                if (scope.editingExisting) {
                    // we shouldn't validate
                    ctrl.$setValidity('validNodeid', true);
                } else {
                    var cleaned = viewValue.replace(/ /g, '').replace(/-/g, '').toUpperCase().trim();
                    if (cleaned.match(/^[A-Z2-7]{52}$/)) {
                        ctrl.$setValidity('validNodeid', true);
                    } else {
                        ctrl.$setValidity('validNodeid', false);
                    }
                }
                return viewValue;
            });
        }
    };
});
