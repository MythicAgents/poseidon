from mythic_payloadtype_container.MythicCommandBase import *
from mythic_payloadtype_container.MythicRPC import *
import base64


class JsImportCallArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="code",
                type=ParameterType.String,
                description="JXA Code to execute from script loaded with jsimport.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        group_name="Default",
                        ui_position=2
                    ),
                ]
            ),
            CommandParameter(
                name="filename",
                display_name="File Registered via 'jsimport' with the function to execute",
                type=ParameterType.ChooseOne,
                dynamic_query_function=self.get_files,
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
        ]

    async def get_files(self, callback: dict) -> [str]:
        file_resp = await MythicRPC().execute("get_file", callback_id=callback["id"],
                                              limit_by_callback=True,
                                              get_contents=False,
                                              filename="",
                                              max_results=-1)
        if file_resp.status == MythicRPCStatus.Success:
            file_names = []
            for f in file_resp.response:
                # await MythicRPC().execute("get_file_contents", agent_file_id=f["agent_file_id"])
                if f["filename"] not in file_names:
                    file_names.append(f["filename"])
            return file_names
        else:
            await MythicRPC().execute("create_event_message", warning=True,
                                      message=f"Failed to get files: {file_resp.error}")
            return []

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
        if "filename" in dictionary and dictionary["filename"] is not None:
            self.add_arg("filename", dictionary["filename"])


class JsImportCallCommand(CommandBase):
    cmd = "jsimport_call"
    needs_admin = False
    help_cmd = 'jsimport_call {  "code": "ObjC.import(\'Cocoa\'); $.NSBeep();" }'
    description = "Execute jxa code from a loaded script via jsimport."
    version = 1
    author = "@its_a_feature_"
    argument_class = JsImportCallArguments
    attackmapping = ["T1059.002"]
    attributes = CommandAttributes(
        # eventually uncomment this once poseidon supports dynamic compilation of commands
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        file_resp = await MythicRPC().execute("get_file", task_id=task.id,
                                              filename=task.args.get_arg("filename"),
                                              limit_by_callback=True,
                                              get_contents=False)
        if file_resp.status == MythicRPCStatus.Success:
            if len(file_resp.response) > 0:
                task.args.add_arg("file_id", file_resp.response[0]["agent_file_id"])
                task.display_params = "code within " + task.args.get_arg("filename")
                task.args.remove_arg("filename")
            elif len(file_resp.response) == 0:
                raise Exception("Failed to find the named file. Have you uploaded it before? Did it get deleted?")
        else:
            raise Exception("Error from Mythic trying to search files:\n" + str(file_resp.error))
        return task

    async def process_response(self, response: AgentResponse):
        pass
