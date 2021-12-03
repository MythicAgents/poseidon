from mythic_payloadtype_container.MythicCommandBase import *
import base64


class JsImportCallArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="code",
                type=ParameterType.String,
                description="JXA Code to execute from script loaded with jsimport.",
            )
        ]

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must provide arguments")
        else:
            try:
                self.load_args_from_json_string(self.command_line)
                self.add_arg(
                    "code", base64.b64encode(self.get_arg("code").encode()).decode()
                )
            except:
                self.add_arg("code", base64.b64encode(self.command_line.encode()).decode())

    async def parse_dictionary(self, dictionary):
        if "code" in dictionary and dictionary["code"] is not None:
            self.add_arg("code", base64.b64encode(dictionary["code"].encode()).decode())


class JsImportCallCommand(CommandBase):
    cmd = "jsimport_call"
    needs_admin = False
    help_cmd = 'jsimport_call {  "code": "ObjC.import(\'Cocoa\'); $.NSBeep();" }'
    description = "Execute jxa code from a loaded script via jsimport."
    version = 1
    author = "@its_a_feature_"
    argument_class = JsImportCallArguments
    attackmapping = ["T1155", "T1064"]
    attributes = CommandAttributes(
        # eventually uncomment this once poseidon supports dynamic compilation of commands
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
