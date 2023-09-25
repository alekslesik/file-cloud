var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

$('.input-file input[type=file]').on('change', function () {
	let file = this.files[0];
	$(this).next().html(file.name);
});

// Function to load file by URL
function downloadFile(url) {
	const link = $('<a>');
	link.attr('href', url);
	link.attr('download', url.split('/').pop());
	$('body').append(link);
	link[0].click();
	link.remove();
}

// Handler when hover
$(document).on('mouseover', function (e) {
	if ($(e.target).hasClass('file')) {
		const fileUrl = $(e.target).data('file-url');
		$(e.target).css('cursor', 'pointer');
		$(e.target).on('click', function () {
			downloadFile(fileUrl);
		});
	}
});

$(document).ready(function () {
    $('#file').change(function () {
        var fileName = $(this).val().split('\\').pop();
        $('#file-placeholder').text(fileName || 'Choose file');
    });

    $('#form').submit(function (event) {
        var fileInput = $('#file')[0];

        if (!fileInput.files || !fileInput.files[0]) {
            event.preventDefault();
            alert('Please choose a file');
        }
    });
});
