var macroApp = angular.module('macroApp', ['ngRoute', 'utils'])
	.config(function($routeProvider) {		
		$routeProvider.when('/mainDashboard', 
			{
				templateUrl:'templates/maindashboard.html',
				controller:'mainDashboardController'
			});
		$routeProvider.when('/history', 
			{
				templateUrl:'templates/history.html',
			});		
		$routeProvider.otherwise({redirectTo:'/mainDashboard'});
	});


macroApp.controller('mainDashboardController', function($scope, calc, http) {
	$scope.formData = {};
	$scope.foodEaten = {};
	$scope.progressBarData = {};
	$scope.gramsEaten = {};
    $scope.date = new Date();  

    // get the current users profile data
    http.getData('GET', 'getUserProfiles', 'json', function(data) {
        
        $scope.userProfiles = data;

        // get dailyi intake and current consumption now that we have the user data
        var params = {'user_id': $scope.userProfiles.id};
	   	http.postData('getUserDailyIntake', params, function(data) {
	   		$scope.foodEaten = data;
	   	});  

        var params = {'user_id': $scope.userProfiles.id};
	   	http.postData('getProgressBarData', params, function(data) {
	   		$scope.progressBarData = data;
	   	});  
    }); 


    http.getData('GET', 'getFoodDropdownData', 'json', function(data) {
        $scope.foodList = data;
    });          

    // add new food to foodprofiles db table and update dropdown
    $scope.addNewFood = function(food) {
    	var params = {
	    	'name': food.name, 
	    	'servingSize': food.servingSize, 
	    	'carbs': food.carbs, 
	    	'protein': food.protein, 
	    	'fat': food.fat
		};

   		http.postData('addNewFood', params, function(response) {
   			if (response == "food already exists") {
   				//trigger alert to see if user wants to update food data
   				console.log(response);
   				$('#duplicateModal').modal('show');
   				return;
   			}

   			triggerAddedPopup("newFoodBtn");

   			console.log(response)
   			$scope.foodList.food.push({	
   			'id': food.id,	
   			'name': food.name, 
	    	'servingSize': food.servingSize, 
	    	'carbs': food.carbs, 
	    	'protein': food.protein, 
	    	'fat': food.fat});

	    	 // clear the form and toggle the dropdown
			$scope.foodForm.$setPristine(true);
			$scope.foodForm.servingSize = "";
			$scope.foodForm.name = "";
			$scope.foodForm.carbs = "";
			$scope.foodForm.protein = "";
			$scope.foodForm.fat = "";
			$scope.addFood = !$scope.addFood; 
		});	
    }

    $scope.addFoodToIntakeList = function(food, user_id) {
    	var params = {
    		'user_id': user_id,
    		"id": food.id, 
	    	'name': food.name, 
	    	'servingSize': food.servingSize, 
	    	'carbs': food.carbs, 
	    	'protein': food.protein, 
	    	'fat': food.fat
		}; 

   		http.postData('addFoodToUserIntake', params, function(response) {
   			// triggerAddedPopup("foodDropdown");

   			if ($scope.foodEaten.food.length == 0) {
   				$scope.foodEaten.food.push({	
		   			'id': food.id,	
		   			'name': food.name, 
			    	'servingSize': food.servingSize, 
			    	'carbs': food.carbs, 
			    	'protein': food.protein, 
			    	'fat': food.fat});
   			} else {
   				$scope.foodEaten.food.unshift({	
		   			'id': food.id,	
		   			'name': food.name, 
			    	'servingSize': food.servingSize, 
			    	'carbs': food.carbs, 
			    	'protein': food.protein, 
			    	'fat': food.fat});
   			}

	    	// update progress bar totals
	    	var params = {'user_id': user_id,'carbs': food.carbs,'protein': food.protein, 'fat': food.fat}
	   		http.postData('updateProgressBars', params, function(response) {
	   			console.log(response)
	   			$scope.progressBarData = response;				
	    	});
    	});
   	}

   	$scope.updateFood = function(food) {
    	var params = {
	    	'name': food.name, 
	    	'servingSize': food.servingSize, 
	    	'carbs': food.carbs, 
	    	'protein': food.protein, 
	    	'fat': food.fat
		};

   		http.postData('updateFood', params, function(response) {   	
   			console.log(response);

   			// update dropdown
   			var listLength = $scope.foodList.food.length;
			for (var i = 0; i < listLength; i++) {

				var test = $scope.foodList.food[i];
				if (test.name == response.name) {
					$scope.foodList.food[i] = response;

					//trigger highlight animation for dropdown
					$("#foodDropdown").addClass("dropdownHighlight");
			        setTimeout(function() {
			            $("#foodDropdown").removeClass("dropdownHighlight");
			        }, 1);
				}
			}
   		});	
   	}		
	
   	// only trigger the added popup if a food was successfully added to the intake
   	// list
   	var triggerAddedPopup = function(popupId) {
   		pid = "#" + popupId;
        
        $(pid).popover('show');
        setTimeout(function() {
            $(pid).popover('hide');
        }, 450);  		
   	}
});	

macroApp.directive('progressBar', function() {
	return {
		restrict: 'E',
		scope: {
			macro: '='
		},
		templateUrl: "/templates/progressBar.html",
		link: function($scope, element, attrs) {

            var watch = $scope.$watch(function() {
                return element.children().length;
            }, function() {
                // Wait for templates to render
                $scope.$evalAsync(function() {
                    // Finally, directives are evaluated
                    // and templates are renderer here
					var id = "#" + $scope.macro.name + "test";		
					bar = new ProgressBar.Circle(id, {
				 	  strokeWidth: 6,
				 	  easing: 'easeInOut',
				 	  duration: 1400,
				 	  //color: '#FFEA82',
				 	  color: '#f0ad4e',
				 	  trailColor: '#eee',
				 	  trailWidth: 1,
				 	  text: {
				   		autoStyleContainer: false,
				 	  },
				 	  svgStyle: null
				 	});			

			  		var stats = ($scope.macro.consumed / $scope.macro.total);
					bar.animate(stats);   
					bar.setText($scope.macro.consumed.toFixed(1) + "g");		                    

					// watch for future changes???
					$scope.$watch('macro', function() {
						var stats = ($scope.macro.consumed / $scope.macro.total);
						bar.animate(stats);   
						bar.setText($scope.macro.consumed.toFixed(1) + "g");
					});

                });
            });
            watch;
    	}		
	}
});