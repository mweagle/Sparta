var SERVICE_NAME = 'TBD'
var cytoscapeView = null

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
  var cloudformationTemplate = null
  try {
    cloudformationTemplate = CLOUDFORMATION_TEMPLATE_JSON
  } catch (e) {
    console.log('Failed to parse template: ' + e.toString())
    cloudformationTemplate = {
      ERROR: e.toString()
    }
  }
  var jsonString = JSON.stringify(cloudformationTemplate, null, 4)
  $('#rawTemplateContent').text(jsonString)
  hljs.initHighlightingOnLoad()
  $('pre code').each(function (i, block) {
    hljs.highlightBlock(block)
  })

  try {
    // Show the cytoscape view
    cytoscapeView = window.cytoscapeView = cytoscape({
      container: $('#cytoscapeDIVTarget'),
      elements: CYTOSCAPE_DATA,
      style: [{
        selector: 'node',
        style: {
          shape: 'round-rectangle',
          content: 'data(label)',
          'background-image': 'data(image)',
          'background-width': '100%',
          'background-height': '100%',
          'background-opacity': '0'
        }
      },
      {
          selector: 'edge',
          style: {
            content: 'data(label)',
            width: 3,
            'curve-style': 'taxi',
            'mid-target-arrow-shape': 'triangle',
            'target-arrow-shape' : 'triangle',
            'line-dash-pattern' : [9, 3],
            'line-fill' : 'linear-gradient',
            'line-gradient-stop-colors' : 'gainsboro slategray black',
            'target-arrow-color' : 'black'
          }
        },
        {
          selector: 'label',
          style: {
            'font-family' : '"Open Sans", Avenir, sans-serif',
            'font-weight' : 'data(labelWeight)'
          },
        }
      ],
      layout: {
        name: 'breadthfirst'
      }
    })
  } catch (err) {
    console.log('Failed to initialize topology view: ' + err)
  }
  var layoutSelectorIDs = [
    '#layout-breadthfirst',
    '#layout-dagre',
    '#layout-cose',
    '#layout-grid',
    '#layout-circle',
    '#layout-concentric'
  ]
  layoutSelectorIDs.forEach(function (eachElement) {
    $(eachElement).click(function (event) {
      event.preventDefault()
      var layoutType = eachElement.split('-').pop()
      console.log('Layout type: ' + layoutType)
      cytoscapeView
        .makeLayout({
          name: layoutType
        })
        .run()
    })
  })
  showView('lambda')
})
