var utils = angular.module('utils', []);	

utils.factory('http', function($http, $q) {
  function handleError(errorData) {
  	console.log(errorData);
  }

  return {
    getData: function(method, url, dataType, successCallback) {
      $http({method: method, url: url, dataType: dataType}).success(successCallback).error(handleError);
    },
    postData: function(url, params, successCallback) {
      $http({url: url, params:params }).success(successCallback).error(handleError);
    }      
  }
});


utils.factory('calc', function() {
  return {
    percent: function(gramsEaten, totalGrams) {
      // make sure we're not dividing by zero
      if (totalGrams == 0) {
        return 0;
      } 

      var percent = (gramsEaten / totalGrams).toFixed(2);
      console.log(percent);
      return percent;
    }
  }
});