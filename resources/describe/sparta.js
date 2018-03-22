var SERVICE_NAME = 'SpartaHTMLAuth-mweagle'

var golangFunctionName = function (cloudFormationResourceName, cloudFormationResources) {
  var res = cloudFormationResources[cloudFormationResourceName] || {}
  var metadata = res.Metadata || {}
  return metadata.golangFunc || 'N/A'
}

var accumulateResources = function (node, pathPart, cloudFormationResources, accumulator) {
  accumulator = accumulator || {}
  var pathPart = pathPart || ''
  var apiResources = node.APIResources || {}
  var apiKeys = Object.keys(apiResources)
  apiKeys.forEach(function (eachKey) {
    var apiDef = apiResources[eachKey]
    var golangName = golangFunctionName(eachKey, cloudFormationResources)
    var resourcePath = pathPart
    var divPanel = $('<div />', {
      'class': 'panel panel-default'
    })
    accumulator[resourcePath] = divPanel

    // Create the heading
    var divPanelHeading = $('<div />', {
      'class': 'panel-heading'
    })
    divPanelHeading.appendTo(divPanel)
    var panelHeadingText = resourcePath + ' (' + golangName + ')'
    var row = $('<div />', {
      'class': 'row'
    })
    row.appendTo(divPanelHeading)
    $('<div />', {
      'text': resourcePath,
      'class': 'col-md-4 text-left'
    }).appendTo(row)
    var golangDiv = $('<div />', {
      'class': 'col-md-8 text-right'
    })
    golangDiv.appendTo(row)
    $('<em />', {
      'text': golangName
    }).appendTo(golangDiv)

    // Create the body
    var divPanelBody = $('<div />', {
      'class': 'panel-body'
    })
    divPanelBody.appendTo(divPanel)

    // Create the method table that will list the METHOD->overview
    var methodTable = $('<table />', {
      'class': 'table table-bordered table-condensed'
    })
    methodTable.appendTo(divPanelBody)

    // Create rows for each method
    var tbody = $('<tbody />')
    tbody.appendTo(methodTable)

    var methods = apiDef.Methods || {}
    var methodKeys = Object.keys(methods)
    methodKeys.forEach(function (eachMethod) {
      var methodDef = methods[eachMethod]

      var methodRow = $('<tr />')
      methodRow.appendTo(tbody)

      // Method
      var methodColumn = $('<td/>', {})
      methodColumn.appendTo(methodRow)
      var methodName = $('<strong/>', {
        text: eachMethod
      })
      methodName.appendTo(methodColumn)
      // Data
      var dataColumn = $('<td/>', {})
      dataColumn.appendTo(methodRow)
      var preElement = $('<pre/>', {})
      preElement.appendTo(dataColumn)
      var codeColumn = $('<code/>', {
        'class': 'JSON',
        text: JSON.stringify(methodDef, null, ' ')
      })
      codeColumn.appendTo(preElement)
    })
  })
  // Descend into children
  var children = node.Children || {}
  var childKeys = Object.keys(children)
  childKeys.forEach(function (eachKey) {
    var eachChild = (children[eachKey])
    accumulateResources(eachChild, pathPart + '/' + eachChild.PathComponent, cloudFormationResources,
      accumulator)
  })
}

function showView (newElementID) {
  $('#view-container').children().hide()
  $('#navBarItems').children().removeClass('active')
  // Set the tab active, the view active
  viewID = '#' + newElementID + '-view'
  tabID = '#' + newElementID + '-tab'
  $(viewID).show()
  $(tabID).addClass('active')
}

$(document).ready(function () {
  console.log('Parsing template')
  var cloudformationTemplate = null
  try {
    cloudformationTemplate = JSON.parse(CLOUDFORMATION_TEMPLATE_RAW)
  } catch (e) {
    console.log('Failed to parse template: ' + e.toString())
    cloudformationTemplate = {
      ERROR: e.toString()
    }
  }
  var jsonString = JSON.stringify(cloudformationTemplate, null, 4);
  $('#rawTemplateContent').text(jsonString)
  hljs.initHighlightingOnLoad();
  $('pre code').each(function (i, block) {
    hljs.highlightBlock(block)
  })
  showView('lambda')
})
