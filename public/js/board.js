var $Colors 	= ["red", "olive", "teal", "grey"];
var $Icons  	= ["stop", "stop", "undo", "play"];
var $Labels 	= ["Finalizar", "Finalizar", "Reactivar", "Activar"];
var $Priority 	= ["", "Alta", "Urgente"];
var $PColors 	= ["", "purple", "orange"];

	
var $ListObject = $(`
	<div class="three wide column list_" data-id="" data-index="">
		<div class="ui top attached borderless compact inverted small menu">
			<div class="ui sub header list title item"></div>
			<div class="right menu">
				<a class="fitted item">
					<div class="ui black compact icon buttons">
						<button class="ui button edit list"><i class="setting icon"></i></button>
						<button class="ui button new task"><i class="plus icon"></i></button>
						<button class="ui button delete list"><i class="close icon"></i></button>
					</div>
				</a>
			</div>
		</div>

		<ul class="ui bottom attached centered secondary segment holder">
			<form class="ui mini hidden form disabled">
				<div class="field">
					<textarea rows="2" placeholder="Descipcion de la tarjeta"/>
				</div>
				<div class="ui fluid mini black button add task">Añadir</div>
			</form>
		</ul>
	</div>
`);

var $CardObject = $(`
	<li class="ui card"  data-id="" data-list="" data-index="" data-state="" data-assignee="" data-duration="" data-activated="">
		<div class="ui top attached mini icon buttons">
			<button class="ui button edit task" data-tooltip="Editar" data-position="bottom center"><i class="setting icon"></i></button>
			<button class="ui button set duration" data-tooltip="Duración" data-position="bottom center"><i class="alarm icon"></i></button>
			<button class="ui button change state" data-tooltip="" data-position="bottom center"><i class="icon"></i></button>
			<button class="ui button delete task" data-tooltip="Eliminar" data-position="bottom center"><i class="close icon"></i></button>
		</div>
		<div class="content">
			<div class="left floated author assignee"></div>
			<div class="right floated meta time duration"></div>
			<div class="description">
				<p></p>
			</div>
		</div>
	</li>
	`);

var $Card;
var $List;


$(function(){
	$(document).ready(function($) {
		$(".board").sortable({
			items: ".column:not(.excluded)",
		});
		$(".board > .list > .holder > form > .field" ).disableSelection();
		$(".holder").sortable({
			items: ".card", 
			connectWith: ".holder", 
			cancel: ".disabled",
		});
		$(".holder").disableSelection();

		$(".new.task").click(function(){
			$(this).closest(".list_").find("form").toggleClass("hidden");;
		});

		$('.message .close')
		  .on('click', function() {
		    $(this)
		      .closest('.message')
		      .transition('fade')
		    ;
		  })
		;

		$(".card").each(function(index, element){
			var $State = $(this).data("state");
			var $CPriority = $(this).data("priority");
			$(this).addClass($Colors[$State]);
			$(this).find(".change.state").attr("data-tooltip", $Labels[$State]);
			$(this).find(".change.state > .icon").addClass($Icons[$State]);
			console.log($CPriority)
			if($CPriority > 0) {
				$(this).find(".priority").html("<a class='ui "+ $PColors[$CPriority] +" empty circular label'></a>&nbsp;"+ $Priority[$CPriority] +"</span>");
			}
			UpdateCardTimeLabel($(this));
			ValidateDuration($(this));
		});

		RefreshCardTime();
	});
});


function ValidateDuration($this){
	if( ($RepoInitTime * 1000) + ($($this).data("duration") ) > ($RepoLimitTime * 1000) ) {
		if($($this).find(".outdate").length == 0) {
			if($($this).find(".extra.content").length) {
				$($this).find(".extra.content").append("<span class='right floated icon outdate'><i class='warning sign icon' />Tarjeta Fuera de Limite</span>");
			} else {
				$($this).append("<div class='extra content'><span class='right floated icon outdate'><i class='warning sign icon' />Tarjeta Fuera de Limite</span></div>");
			}
		}
	} else {
		if($($this).find(".priority").length) {
			$($this).find(".outdate").remove();
		} else {
			$($this).find(".extra.content").remove();
		}
	}
}

function RefreshCardTime(){
	console.log("doing time update")
	$.each($(".card"), function(index, element) {
		if($(this).data("state") == 1){
			UpdateCardTimeLabel($(this));
		}
	});

	setTimeout(function(){
		RefreshCardTime();
	}, 60000);
}


function UpdateCardTimeLabel($this){
	switch ($($this).data("state")) {
		case 0:
			$($this).find(".meta.duration").text("Tarjeta Expirada");
			break;
		case 1:
			if($($this).data("duration") > 0){
				var $Time = GetRemaningTime($($this).data("duration"), $($this).data("activated"));
				if(($Time.Days + $Time.Hours + $Time.Minutes) > 0) {
					FormatTimeLabel($($this), $Time);
				} else {
					ExpireCard($this);
				}
			} else {
				$($this).find(".meta.duration").text("");
			}
			break;
		case 2:
			$($this).addClass("disabled");
			$($this).find(".meta.duration").text("Tarjeta Completada");
			break;
		default:
			if($($this).data("duration") > 0){
				var $Units = GetTimeUnits($($this).data("duration") / 1000);
				FormatTimeLabel($($this), $Units);
			}
			break;
	}
}

function FormatTimeLabel($this, Time) {
	var Label = ((Time.Days + Time.Hours + Time.Minutes) > 1) ? "Restan " : "Resta ";
	if(Time.Days > 0) {
		Label = (Time.Days > 1) ? Label + Time.Days + " Dias" : Label + Time.Days + " Dia";
		if(Time.Hours > 0 || Time.Minutes > 0){
			Label = (Time.Hours > 0 && Time.Minutes > 0) ? Label + ", " : Label + " y ";
		}
	}

	if(Time.Hours > 0) {
		Label +=  Time.Hours + "h";
		if(Time.Minutes > 0){
			Label += " y ";
		}
	}

	if(Time.Minutes > 0) {
		Label += Time.Minutes + "m";
	}
	$($this).find(".meta.duration").text(Label);
}



$('#NewList').modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onApprove: function(){
		$.ajax({
			url: "/api/v1" + $RepoLink + "/board/list",
			type: 'POST',
			dataType: 'JSON',
			data: {title: $("#Title").val(), index: $(".list_").length,},
		})
		.done(function(data, textStatus, xhr) {
			var $NewList = $ListObject.clone(true);
			$($NewList).data("id", data.ID);
			$($NewList).data("index", data.Index);
			$($NewList).find(".list.title").text(data.Title);
			
			$NewList.find(".edit.list")[0].addEventListener("click", function(button){
				$List = $(this).closest(".list_");
				$("#EditList").modal("show");
			}, false)

			$NewList.find(".delete.list")[0].addEventListener("click", function(button){
				$List = $(this).closest(".list_");
				$("#DeleteList").modal("show");
			}, false)

			$NewList.find(".new.task")[0].addEventListener("click", function(){
				$(this).closest(".list_").find("form").toggleClass("hidden");
			}, false);

			$NewList.find('.add.task')[0].addEventListener("click", function(button){
				NewCard($($(button.target).closest(".list_"))[0]);
			}, false)
			
			$(".column.disabled").before($NewList);
			$(".board").sortable({
				items: ".column:not(.excluded)",
			});
		})
		.fail(function(data, textStatus, xhr) {
			console.log(data);
			console.log("error");
		});
	},
	onHide: function(){
		$("#NewListForm")[0].reset();
	},
});

$("#NewListForm").submit(function(event){
	event.preventDefault();
	return false
});

$(".add.list").click(function(){
	$("#NewList").modal("show");
});


$('#EditList').modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onShow: function(){
		$("#EditTitle").val($($List).find(".header").text());
	},
	onApprove: function() {
		UpdateList($List);
	},
	onHide: function(){
		$("#EditList").find("form")[0].reset();
	},
});

function UpdateList($this){
	console.log("updating");
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/list/" + $($this).data("id"),
		type: 'PATCH',
		dataType: 'JSON',
		data: {title: $("#EditTitle").val(), index: $($this).index()}
	})
	.done(function(data, textStatus, xhr) {
		$($this).data("index", data.Index)
		$($this).find('.list.title').text(data.Title);
	})
	.fail(function(data, textStatus, xhr) {
		console.log(data);
		console.log("error");
	});
}

$(".edit.list").click( function(event) {
	$List = $(this).closest(".list_");
	$("#EditList").modal("show");
});


$("#DeleteList").modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onShow	 : function(){
		$("#DeleteList").find(".grouped.fields").html("");
		if($($List).find(".card").length){
			$("#DeleteList").find(".delete.message").html("<p>Esta operacion es irreversible asegurate de que quieres borrar las siguientes tarjetas:</p>")
			if($($List).find(".card.olive").length){
				$("#DeleteList").find(".grouped.fields").append($(`
					<div class="field">
						<div class="ui checkbox">
						  <input type="checkbox">
						  <label><div class="ui olive empty circular label"></div> Se eliminaran `+ $($List).find(".card.olive").length +` tarjeta(s) activa(s)</label>
						</div>
					</div>
				`));
			}

			if($($List).find(".card.teal").length){
				$("#DeleteList").find(".grouped.fields").append($(`
					<div class="field">
						<div class="ui checkbox">
						  <input type="checkbox">
						  <label><div class="ui teal empty circular label"></div> Se eliminaran `+ $($List).find(".card.teal").length +` tarjeta(s) finalizada(s)</label>
						</div>
					</div>
				`));
			}

			if($($List).find(".card.red").length){
				$("#DeleteList").find(".grouped.fields").append($(`
					<div class="field">
						<div class="ui checkbox">
						  <input type="checkbox">
						  <label><div class="ui red empty circular label"></div> Se eliminaran `+ $($List).find(".card.red").length +` tarjeta(s) pendiente(s)</label>
						</div>
					</div>
				`));
			}

			if($($List).find(".card.grey").length){
				$("#DeleteList").find(".grouped.fields").append($(`
					<div class="field">
						<div class="ui checkbox">
						  <input type="checkbox">
						  <label><div class="ui grey empty circular label"></div> Se eliminaran `+ $($List).find(".card.grey").length +` tarjeta(s) planeada(s)</label>
						</div>
					</div>
				`));
			}

			$("ui.checkbox").checkbox();
		} else {
			$("#DeleteList").find(".delete.message").html("<p>Esta lista no contiene tarjetas</p>")
		}
		
	},
	onApprove : function(){
		if($("#DeleteList").find(".ui.checkbox").length){
			var $Checked = true; 
			$.each($(".ui.checkbox"), function(index, element){
					$Checked = $Checked && $(this).checkbox("is checked");
			});
			if(!$Checked){return false}
		}
		DeleteList($List);
	},
});

function DeleteList($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/list/" + $($this).data("id"),
		type: 'DELETE',
		dataType: 'JSON',
	})
	.done(function(data, textStatus, xhr) {
		var $Index = $($this).index();
		$($this).remove();
		$(".list_").each(function(index, element){
			if($(this).data("index") >= $Index){
				UpdateList($(this));
			}
		});
	})
	.fail(function(data, textStatus, xhr) {
		console.log("error");
	});
}

$(".delete.list").click(function(event) {
	$List = $(this).closest(".list_");
	$("#DeleteList").modal("show");		
});

	

$( ".board" ).on( "sortupdate", function(event, ui){
	if($(ui.item.context).hasClass('list_')){
		var $LastIndex	= $(ui.item.context).data("index");
		var $Index 		= $(ui.item.context).index();

		if(($Index + $LastIndex) == ($(".list_").length - 1)){
			$.each($(".list_"), function(index, element) {
				UpdateList($(this));
			});
		} else if($Index > $LastIndex){
			$.each($(".list_"), function(index, element) {
				if($(this).index() >= $LastIndex && $(this).index() <= $Index){
					UpdateList($(this));
				}
			});
			
		} else {
			$.each($(".list_"), function(index, element) {
				if($(this).index() >= $Index && $(this).index() <= $LastIndex){
					UpdateList($(this));
				}
			});
		}
	}
});


$(".holder").on('sortupdate', function(event, ui) {
	SortInsde(ui)
});

function SortInsde($this) {
	var $Container 	= $($this.item.context).closest(".list_");
	var $LastIndex 	= $($this.item.context).data("index");
	var $Index 	 	= $($this.item.context).index();

	if($this.sender == null && $($Container).data("id") == $($this.item.context).data("list")){
		if(($Index + $LastIndex) == ($($Container).find(".card").length + 1)){
			$.each($($Container).find(".card"), function(index, element) {
				MoveCard($(this));
			});
		} else if($Index > $LastIndex){
			$.each($($Container).find(".card"), function(index, element) {
				if($(this).index() >= $LastIndex && $(this).index() <= $Index){
					MoveCard($(this));
				}
			});
			
		} else {
			$.each($($Container).find(".card"), function(index, element) {
				if($(this).index() >= $Index && $(this).index() <= $LastIndex){
					MoveCard($(this));
				}
			});
		}
	}
}

$(".holder").on("sortreceive", function(event, ui) {
	SortAdd(ui)
});

function SortAdd($this) {
	var $Sender  = $($this.sender.context).closest(".list_");
	var $Reciver = $($this.item.context).closest(".list_");

	TransferCard($(".card[data-id=" + $($this.item.context).data("id") + "]"), $($Reciver).data("id"));

	$.each($($Sender).find(".card"), function(index, element) {
		if($(this).index() >= $($this.item.context).data("index")){
			MoveCard($(this));
		}
	});

	$.each($($Reciver).find(".card"), function(index, element) {
		if($(this).index() >= $($this.item.context).index()){
			MoveCard($(this));
		}
	});
}

$(".holder").on("sort", function(event, ui){
	SortCancel(ui)
});

function SortCancel($this){
	if($($this.item.context).hasClass("disabled")){
		$(this).sortable("cancel");
	}
}

function MoveCard($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/move/" + $($this).data("id"),
		type: 'PATCH',
		dataType: 'JSON',
		data: {index: $($this).index()},
	})
	.done(function(data, textStatus, xhr) {
		$($this).data("index", data.Index);
	})
	.fail(function(data, textStatus, xhr) {
		console.log(textStatus, xhr);
		console.log("error moving");
	});
}

function TransferCard($this, $Destination){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/move/" + $($this).data("id"),
		type: 'POST',
		async: false,
		dataType: 'JSON',
		data: {list: $Destination},
	})
	.done(function(data, textStatus, xhr) {
		$($this).data("list", data.List);
	})
	.fail(function(data, textStatus, xhr) {
		console.log(textStatus, xhr);
		console.log("error transfering");
	});
}



function NewCard($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card",
		type: 'POST',
		dataType: 'JSON',
		data: {list: $($this).data("id"), body: $($this).find("textarea").val(), index: $($this).find(".card").length + 1,},
	})
	.done(function(data, textStatus, xhr) {
		var $NewCard = $CardObject.clone(true);
		$($NewCard).attr("data-id", data.ID);
		$($NewCard).data("id", data.ID);
		$($NewCard).data("list", data.List);
		$($NewCard).data("index", data.Index);
		$($NewCard).data("state", data.State);
		$($NewCard).data("priority", data.Priority);
		$($NewCard).data("duration", data.Duration);
		$($NewCard).data("activated",data.Activated);

		$($NewCard).addClass($Colors[data.State]);

		$($NewCard).find('p').text(data.Body);

		$NewCard.find('.edit.task')[0].addEventListener("click", function(button){
			$Card = $(this).closest(".card");
			$("#EditCard").modal("show");
		}, false);

		$NewCard.find(".set.duration")[0].addEventListener("click", function(button){
			$Card = $(this).closest(".card");
			$("#EditDuration").modal("show");
		}, false);

		$NewCard.find(".change.state > i").addClass($Icons[data.State]);
		$NewCard.find(".change.state").attr("data-tooltip", $Labels[data.State]);

		$NewCard.find(".change.state")[0].addEventListener("click", function(button){
			UpdateCardState($(this).closest(".card"));
		}, false);

		$NewCard.find('.delete.task')[0].addEventListener("click", function(button){
			$Card = $(this).closest(".card");
			$("#DeleteCard").modal("show");
		}, false);
		$($this).find(".holder").append($NewCard);
		$($this).find("textarea").val("");
		$(".holder").sortable({
			items: ".card", 
			connectWith: ".holder", 
			cancel: ".disabled",
		});
		$(".holder").on('sortupdate', function(event, ui) {
			SortInsde(ui)
		});


		$(".holder").on("sortreceive", function(event, ui) {
			SortAdd(ui)
		});

		$(".holder").on("sort", function(event, ui){
			SortCancel(ui)
		});
	})
	.fail(function(data, textStatus, xhr) {
		console.log(data);
		console.log("error");
	});
}

$(".add.task").click(function(){
	NewCard($(this).closest(".list_"));
});


$('#EditCard').modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onShow	 : function(){
		var $Assignee 	= $($Card).data("assignee");

		$("#Description").val($($Card).find("p").text());
		if($Assignee.length){
			$("#Assignee").dropdown("set selected", $Assignee);
		}
		$("#Priority").dropdown("set selected", $($Card).data("priority"))
	},
	onApprove : function(){
		UpdateCard($Card)
	},
	onHide: function(){
		$("#Description").val("")
		$("#Assignee").dropdown("clear");
		$("#Priority").dropdown("clear");
	},
});

function UpdateCard($this){
	console.log($("#Priority").dropdown("get value"))
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/" + $($this).data("id"),
		type: 'PATCH',
		dataType: 'JSON',
		data: {body: $("#Description").val(), assignee: $("#Assignee").dropdown("get value"), priority: $("#Priority").dropdown("get value"),}
	})
	.done(function(data, textStatus, xhr) {
		$($this).find("p").text(data.Body);
		if(data.Assignee != null) {
			if($($this).data("assignee") != data.Assignee.username){
				$($this).data("assignee", data.Assignee.username);
				$($this).find(".assignee").html("<img class='ui avatar image' src='" + data.Assignee.avatar_url + "'>" + data.Assignee.username);
			}
		}
		$($this).data("priority", data.Priority);
		if(data.Priority > 0) {
			if($($this).find(".extra.content").length) {
				$($this).find(".priority").html("<a class='ui "+ $PColors[data.Priority] +" empty circular label'></a>"+ $Priority[data.Priority] +"</span>");
			} else {
				$($this).append("<div class='extra content'><a class='ui "+ $PColors[data.Priority] +" empty circular label'></a>&nbsp;"+ $Priority[data.Priority] +"</span> </div>");
			}
		} else {
			if($($this).find(".outdate").length == 0){
				$($this).find(".extra.content").remove();
			}
		}
	})
	.fail(function(data, textStatus, xhr) {
		if(data.status == 403){
			$(".message.warning").find("p").text("No tienes permiso para editar esta tarjeta");
			$(".message.warning").transition({animation: "fade", duration: "4s",}).transition("scale");
		}
		console.log("error");
	});
}

$('.edit.task').click(function(event){
	$Card = $(this).closest(".card");
	$("#EditCard").modal("show");
});

function UpdateCardState($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/state/" + $($this).data("id"),
		type: 'POST',
		dataType: 'JSON',
	})
	.done(function(data, textStatus, xhr) {
		$($this).data("activated", data.Activated);
		$($this).data("duration", data.Duration);
		if(data.State == 2){
			$($this).addClass("disabled");
		} else {
			$Card = $($this);
			$($this).removeClass("disabled");
		}
		if($($this).data("state") == 2){
			$("#EditDuration").modal("show");
		}
		$($this).removeClass($Colors[$($this).data("state")]).addClass($Colors[data.State]);
		$($this).find(".change.state > i").removeClass($Icons[$($this).data("state")]).addClass($Icons[data.State]);
		$($this).find(".change.state").attr("data-tooltip", $Labels[data.State]);
		$($this).data("state", data.State);

		UpdateCardTimeLabel($this);
	})
	.fail(function(data, textStatus, xhr) {
		if(data.status == 403){
			$(".message.warning").find("p").text("No tienes permiso para cambiar el estado de esta tarjeta");
			$(".message.warning").transition({animation: "fade", duration: "4s",}).transition("scale");
		}
		console.log("error");
	});
}

$(".change.state").click(function(){
	UpdateCardState($(this).closest(".card"));
});

function ExpireCard($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/expire/" + $($this).data("id"),
		type: 'POST',
		dataType: 'JSON',
	})
	.done(function(data, textStatus, xhr) {
		$($this).removeClass("olive").addClass("red");
		$($this).data("state", data.State);
		UpdateCardTimeLabel($this);
	})
	.fail(function(data, textStatus, xhr) {
		console.log(xhr);
		console.log(data);
	});
}


$('#EditDuration').modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onShow	 : function(){
		var $Units = GetTimeUnits(parseInt($($Card).data("duration")) / 1000);
		$("#Days").val($Units.Days);
		$("#Hours").val($Units.Hours);
		$("#Minutes").val($Units.Minutes);
	},
	onApprove : function(){
		UpdateCardDuration($Card);
	},
	onHide: function(){
		$("#DurationForm")[0].reset();
	},
});


function UpdateCardDuration($this){
	var $DD = $.isNumeric($("#Days").val())    ?  parseInt($("#Days").val())    : 0;
	var $HH = $.isNumeric($("#Hours").val())   && parseInt($("#Hours").val())   < 12 ? parseInt($("#Hours").val())   : 0;
	var $MM = $.isNumeric($("#Minutes").val()) && parseInt($("#Minutes").val()) < 60 ? parseInt($("#Minutes").val()) : 0;
	var $Duration = GetTime($DD, $HH, $MM);
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/duration/" + $($this).data("id"),
		type: 'POST',
		dataType: 'JSON',
		data: {duration: $Duration}
	})
	.done(function(data, textStatus, xhr) {
		$($this).data("duration", data.Duration);
		$($this).data("activated", data.Activated);
		UpdateCardTimeLabel($this);
		if($($this).data("state") == 0){
			UpdateCardState($this);
		}
		ValidateDuration($this);
	})
	.fail(function(data, textStatus, xhr) {
		if(data.status == 403){
			$(".message.warning").find("p").text("No tienes permiso para editar esta tarjeta");
			$(".message.warning").transition({animation: "fade", duration: "4s",}).transition("scale");
		}
		console.log("error");
	});
}

$('.set.duration').click(function(event){
	$Card = $(this).closest(".card");
	$("#EditDuration").modal("show");
});


function GetTime($Days, $Hours, $Minutes){
	return (($Days * 86400) + ($Hours * 3600) + ($Minutes * 60)) * 1000;
}

function GetTimeUnits($Time) {
	var $Units = {};

	// calculate (and subtract) whole days
	var $Days = Math.floor($Time / 86400);
	$Units.Days = $Days;
	$Time -= $Days * 86400;

	// calculate (and subtract) whole hours
	var $Hours = Math.floor($Time / 3600) % 24;
	$Units.Hours = $Hours;
	$Time -= $Hours * 3600;

	// calculate (and subtract) whole minutes
	var $Minutes = Math.floor($Time / 60) % 60;
	$Units.Minutes = $Minutes;
	$Time -= $Minutes * 60;

	return $Units;
}

function GetRemaningTime($Time, $ActivatedDate){
	var $LimitDate 	= ($ActivatedDate * 1000) + $Time;
	var $Seconds 	= $LimitDate > Date.now() ? ($LimitDate - Date.now()) / 1000 : 0;
	return GetTimeUnits($Seconds);
}	

$("#DeleteCard").modal({
	approve  : '.positive, .approve, .ok',
	deny     : '.negative, .deny, .cancel',
	onShow	 : function(){
		$("span.target").append($($Card).clone());
	},
	onApprove : function(){
		DeleteCard($Card)
	},
	onHide: function(){
		$("span.target").html("");
	}
});

function DeleteCard($this){
	$.ajax({
		url: "/api/v1" + $RepoLink + "/board/card/" + $($this).data("id"),
		type: 'DELETE',
	})
	.done(function(data, textStatus, xhr) {
		$($this).remove();
		$($this).closest(".list_").find(".card").each(function(index, element){
			if($(this).index() >= $($this).index()){
				UpdateCard($(this));
			}
		});
	})
	.fail(function(data, textStatus, xhr) {
		if(data.status == 403){
			$(".message.warning").find("p").text("No tienes permiso para borrar esta tarjeta");
			$(".message.warning").transition({animation: "fade", duration: "4s",}).transition("scale");
		}
		console.log(textStatus);
	});
}

$(".delete.task").click(function(event) {
	$Card = $(this).closest(".card");
	$("#DeleteCard").modal("show");
});
