$(document).ready(function () {
  // Attach a click event listener to the login button
  $("#login_btn").click(function () {
      $("#secondary_body").attr("action", "/login/login_submit/");
  });

  // Attach a click event listener to the signup button
  $("#signup_btn").click(function (event) {
      var username = $("#username_input").val();
      var password = $("#password_input").val();
      var isValid = true; // Assume all inputs are valid initially

      if (username.length < 4) {
          $("#username_input")
              .addClass("invalid-placeholder")
              .css("color", "red")
              .focus(function () {
                  $(this)
                      .removeClass("invalid-placeholder")
                      .css("color", "black");
              });
          isValid = false; // Input is invalid
      }

      if (password.length < 8) {
          $("#password_input")
              .addClass("invalid-placeholder")
              .css("color", "red")
              .focus(function () {
                  $(this)
                      .removeClass("invalid-placeholder")
                      .css("color", "black");
              });
          isValid = false; // Input is invalid
      }

      // Only proceed if all inputs are valid
      if (!isValid) {
          event.preventDefault(); // Prevent form submission
      } else {
          $("#secondary_body").attr("action", "/login/signup_submit/").submit();
      }
  });
});
