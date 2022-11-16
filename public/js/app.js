$(function () {

    $("#upload_link").click(function (e) {
        e.preventDefault();
        var text = prompt("File Upload", "Enter a nice caption here...");
        $("#caption").val(text);
        $("#upload").trigger('click');
    });

    $("span.img-action.heart").click(function (e) {
        var id = $(this).data("id");
        $(this).find(">:first-child").addClass("redClass");

        $.ajax({
            url: `/photos/${id}/like`,
            type: 'POST'
        }).done(function (data) {
            $("#likeCount").text(data.likes + " likes")
        }).fail(function (jqXHR, textStatus) {
            console.log("An error occurred: " + textStatus);
        });

    });

    $("span.img-action.comment").click(function () {
        var id = $(this).data("id");
        $(`#comment-${id}`).focus();
    });

    $("span.img-action.trash").click(function (e) {
        e.preventDefault();
        if (window.confirm("Are you sure?")) {
            var id = $(this).data("id");

            $.ajax({
                url: `/photos/${id}`,
                type: 'DELETE'
            }).done(function (data) {
                window.location.replace('/photos/');
            }).fail(function (jqXHR, textStatus) {
                console.log("An error occurred: " + textStatus);
            });
        }
    });

    $("input.comment").keypress(function (e) {
        var id = $(this).data("id");
        var input = $(this)
        var comment = $(this).val()
        if (e.which == 13) {
            $.ajax({
                url: `/photos/${id}/comment`,
                type: 'POST',
                dataType: 'json',
                data: '{"comment": "' + comment + '"}'
            }).done(function (data) {
                console.log("Posted comment: " + comment);
                $("#photoBody").append('<p><b>' + data.username + '</b>&nbsp;<span class="text-muted">' + comment + '</span></p>');
                input.val("")
            }).fail(function (jqXHR, textStatus) {
                console.log("An error occurred: " + textStatus);
            });
        }
    });

    $("#follow").click(function (e) {
        var id = $(this).data("id");
        var followbtn = $(this)

        $.ajax({
            url: `/user/${id}/follow`,
            type: 'POST'
        }).done(function (data) {
            $(followbtn).hide()
            $("#unfollow").show()
        }).fail(function (jqXHR, textStatus) {
            console.log("An error occurred: " + textStatus);
        });

    });

    $("#unfollow").click(function (e) {
        var id = $(this).data("id");
        var unfollowbtn = $(this)

        $.ajax({
            url: `/user/${id}/unfollow`,
            type: 'POST'
        }).done(function (data) {
            unfollowbtn.hide()
            $("#follow").show()
        }).fail(function (jqXHR, textStatus) {
            console.log("An error occurred: " + textStatus);
        });

    });

});
