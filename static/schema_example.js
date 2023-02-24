
      // Initialize the editor with a JSON schema
      var editor = new JSONEditor(document.getElementById('editor_holder'),{
        schema: {
          type: "object",
          title: "Text",
          required: ["fontSize","color","font","weight","possibleFonts"],
          properties: {
            fontSize: {
              format: "choices",
              type: "integer",
              enum: [10,11,12,14,16,18,20,22,24,36,48,60,72,100],
              default: 24,
              options: {
                choices_options: {
                  shouldSort: false
                }
              }
            },
            color: {
              type: "string",
              format: "choices",
              enum: ["black","red","green","blue","yellow","orange","purple","brown","white","cyan","magenta"]
            },
            font: {
              type: "string",
              format: "choices",
              enumSource: "possi8
                "possible_fonts": "root.possibleFonts"
              }
            },
            weight: {
              type: "string",
              format: "choices",
              enum: ["normal","bold","bolder","lighter"],
              options: {
                enum_titles: ["Normal (400)","Bold (700)","Bolder (900)","Lighter (200)"]
              }
            },
            possibleFonts: {
              type: "array",
              format: "choices",
              uniqueItems: true,
              items: {
                type: "string"
              },
              default: ["Arial","Times","Helvetica","Comic Sans"]
            }
          }
        },
        startval: {
          color: "red"
        }
      });

      // Hook up the submit button to log to the console
      document.getElementById('submit').addEventListener('click',function() {
        // Get the value from the editor
        console.log(editor.getValue());
      });
    