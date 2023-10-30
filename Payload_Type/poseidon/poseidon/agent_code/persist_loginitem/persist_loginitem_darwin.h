//
//  persist_loginitem.h
//  loginitempersisttest
//
//  Created by afrosamurai on 8/19/21.
//  Copyright © 2021 Test. All rights reserved.
//

#ifndef persist_loginitem_h
#define persist_loginitem_h

extern const char* removeSessionLoginItems(char* removePath);
extern const char* removeGlobalLoginItems(char* removePath, char* removeName);
extern const char* addGlobalLoginItem(unsigned char* path, unsigned char* name);
extern const char* addSessionLoginItem(unsigned char* path, unsigned char* name);
extern const char* listSessionLoginItems();
extern const char* listGlobalLoginItems();
extern const char * removeitem(char* path, char* name);
extern const char * addsessionitem(char* path, char* name);
extern const char * addglobalitem(char* path, char* name);
extern const char * listitems();

#endif /* persist_loginitem_h */