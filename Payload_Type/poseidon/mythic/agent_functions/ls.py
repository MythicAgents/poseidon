from mythic_payloadtype_container.MythicCommandBase import *
import json


class LsArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        self.add_arg("file_browser", False, type=ParameterType.Boolean)
        if len(self.command_line) > 0:
            try:
                # will get a JSON input if the tasking comes from the file browser
                tmp_json = json.loads(self.command_line)
                self.command_line = tmp_json["path"] + "/" + tmp_json["file"]
                self.add_arg("file_browser", True, type=ParameterType.Boolean)
            except:
                pass
            self.add_arg("path", self.command_line)
        else:
            self.add_arg("path", ".")

    async def parse_dictionary(self, dictionary):
        if "path" in dictionary and "file" in dictionary:
            self.add_arg("file_browser", value=True, type=ParameterType.Boolean)
            self.add_arg("path", value=dictionary["path"] + "/" + dictionary["file"])
        else:
            self.add_arg("file_browser", value=False, type=ParameterType.Boolean)
            self.add_arg("path", value=".")


class LsCommand(CommandBase):
    cmd = "ls"
    needs_admin = False
    help_cmd = "ls [directory]"
    description = "List directory."
    version = 1
    supported_ui_features = ["file_browser:list"]
    author = "@xorrior"
    argument_class = LsArguments
    attackmapping = ["T1083"]
    browser_script = BrowserScript(script_name="ls", author="@its_a_feature_")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        if task.args.has_arg("file_browser") and task.args.get_arg("file_browser"):
            host = task.callback.host
            task.display_params = host + ":" + task.args.get_arg("path")
        else:
            task.display_params = task.args.get_arg("path")
        return task

    async def process_response(self, response: AgentResponse):
        pass
