mciModule.config(function($routeProvider, $locationProvider) {
  function applyProject(_, url) {
    //FIXME (minor) $window might be a more idiomatic choice
    //      $window could not be used in $config
    return url + '/' + window.project
  }

  $routeProvider
    // TODO make a mixin for this
    // Redirects to the same URL with filled in project id
    .when('/perf-bb/change-points', {
      redirectTo: applyProject,
    })
    .when('/perf-bb/change-points/:projectId', {
      templateUrl: '/static/app/perf-bb/change-points.html',
      controller: 'SignalProcessingCtrl',
      controllerAs: 'spvm',
    })
    // TODO make a mixin for this
    // Redirects to the same URL with filled in project id
    .when('/perf-bb/failures', {
      redirectTo: applyProject,
    })
    .when('/perf-bb/failures/:projectId', {
      templateUrl: '/static/app/perf-bb/failures.html',
      controller: 'PerfBBFailuresCtrl',
      controllerAs: 'pfvm',
    })
    // This was added for compatibility with server-side routing
    .otherwise({
      redirectTo: function(_, url) {
        // When UI route doesn't exist, redirect user to the URL
        window.location = url
      }
    })
  // Enable client-side routing and history
  $locationProvider.html5Mode({enabled: true, requireBase: false})
})
